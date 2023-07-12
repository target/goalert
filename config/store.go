package config

import (
	"context"
	cryptoRand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/jsonutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// Store handles saving and loading configuration from a postgres database.
type Store struct {
	rawCfg       Config
	cfgVers      int
	fallbackURL  string
	explicitURL  string
	mx           sync.RWMutex
	db           *sql.DB
	keys         keyring.Keys
	latestConfig *sql.Stmt
	setConfig    *sql.Stmt
	lock         *sql.Stmt

	closeCh chan struct{}
}

type StoreConfig struct {
	DB   *sql.DB
	Keys keyring.Keys

	// FallbackURL is the URL to use when the DB config does not specify a public URL.
	FallbackURL string

	// ExplicitURL is the full public URL to use for all links.
	ExplicitURL string

	// IngressEmailDomain is the domain to use for ingress email addresses.
	IngressEmailDomain string
}

// NewStore will create a new Store with the given StoreConfig parameters. It will automatically detect
// new configuration changes.
func NewStore(ctx context.Context, cfg StoreConfig) (*Store, error) {
	p := util.Prepare{Ctx: ctx, DB: cfg.DB}

	s := &Store{
		db:           cfg.DB,
		fallbackURL:  cfg.FallbackURL,
		explicitURL:  cfg.ExplicitURL,
		latestConfig: p.P(`select id, data, schema from config where schema <= $1 order by id desc limit 1`),
		setConfig:    p.P(`insert into config (id, schema, data) values (DEFAULT, $1, $2) returning (id)`),
		lock:         p.P(`lock config in exclusive mode`),
		keys:         cfg.Keys,
		closeCh:      make(chan struct{}),
	}

	s.rawCfg.SMTPServer.EmailDomain = cfg.IngressEmailDomain

	if p.Err != nil {
		return nil, p.Err
	}

	var err error
	permission.SudoContext(ctx, func(ctx context.Context) {
		err = s.Reload(ctx)
	})
	if err != nil {
		return nil, err
	}

	var seed int64
	err = binary.Read(cryptoRand.Reader, binary.BigEndian, &seed)
	if err != nil {
		return nil, err
	}
	src := rand.New(rand.NewSource(
		seed,
	))

	logger := log.FromContext(ctx)
	go func() {
		randDelay := func() time.Duration {
			return 30*time.Second + time.Duration(src.Int63n(int64(30*time.Second)))
		}
		t := time.NewTimer(randDelay())
		for {
			select {
			case <-t.C:
				t.Reset(randDelay())
				permission.SudoContext(logger.BackgroundContext(), func(ctx context.Context) {
					err := s.Reload(ctx)
					if err != nil {
						log.Log(ctx, errors.Wrap(err, "config auto-reload"))
					}
				})
			case s.closeCh <- struct{}{}:
				close(s.closeCh)
				return
			}
		}
	}()

	return s, nil
}

// Shutdown stops the config reloader.
func (s *Store) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.closeCh:
	}
	return nil
}

func wrapTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}

// Reload will re-read and update the current config state from the DB.
func (s *Store) Reload(ctx context.Context) error {
	cfg, id, err := s.reloadTx(ctx, nil)
	if err != nil {
		return err
	}
	rawCfg := *cfg
	rawCfg.fallbackURL = s.fallbackURL
	rawCfg.explicitURL = s.explicitURL

	err = cfg.Validate()
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "validate config"))
	}

	s.mx.Lock()
	oldVers := s.cfgVers
	s.cfgVers = id
	s.rawCfg = rawCfg
	s.mx.Unlock()

	if oldVers != id {
		log.Logf(ctx, "Loaded config version %d ", id)
	}

	return nil
}

// ServeConfig handles requests to read and write the config json.
func (s *Store) ServeConfig(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	switch req.Method {
	case "GET":
		id, _, data, err := s.ConfigData(ctx, nil)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="goalert-config.%d.json"`, id))
		_, _ = w.Write(data)
	case "PUT":
		data, err := io.ReadAll(req.Body)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		id, err := s.SetConfigData(ctx, nil, data)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		log.Logf(ctx, "Set configuration to version %d (schema version %d)", id, SchemaVersion)

		err = s.Reload(ctx)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		w.WriteHeader(204)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// ConfigData will return the current raw config data from the DB.
func (s *Store) ConfigData(ctx context.Context, tx *sql.Tx) (id, schemaVersion int, data []byte, err error) {
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return 0, 0, nil, err
	}

	err = wrapTx(ctx, tx, s.latestConfig).QueryRowContext(ctx, SchemaVersion).Scan(&id, &data, &schemaVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, SchemaVersion, []byte("{}"), nil
	}
	if err != nil {
		return 0, 0, nil, err
	}

	data, _, err = s.keys.Decrypt(data)
	if err != nil {
		return 0, 0, nil, errors.Wrap(err, "decrypt config")
	}

	return id, schemaVersion, data, nil
}

// SetConfigData will replace the current DB config with data.
func (s *Store) SetConfigData(ctx context.Context, tx *sql.Tx, data []byte) (int, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return 0, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return 0, errors.Wrap(err, "validate config")
	}

	data, err = s.keys.Encrypt("CONFIG", data)
	if err != nil {
		return 0, errors.Wrap(err, "encrypt config")
	}

	var id int
	err = wrapTx(ctx, tx, s.setConfig).QueryRowContext(ctx, SchemaVersion, data).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Store) reloadTx(ctx context.Context, tx *sql.Tx) (*Config, int, error) {
	id, schemaVersion, data, err := s.ConfigData(ctx, tx)
	if err != nil {
		return nil, 0, err
	}

	var c Config
	switch schemaVersion {
	case 1:
		err = json.Unmarshal(data, &c)
		if err != nil {
			return nil, 0, errors.Wrap(err, "unmarshal config")
		}
	default:
		return nil, 0, errors.Errorf("invalid config schema version found: %d", schemaVersion)
	}

	c.data = data
	return &c, id, nil
}

// UpdateConfig will update the configuration in the DB and perform an immediate reload.
func (s *Store) UpdateConfig(ctx context.Context, fn func(Config) (Config, error)) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "config: update", tx)

	id, err := s.updateConfigTx(ctx, tx, fn)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Logf(ctx, "Set configuration to version %d (schema version %d)", id, SchemaVersion)

	return s.Reload(ctx)
}

// SetConfig will replace the configuration in the DB and perform an immediate reload.
func (s *Store) SetConfig(ctx context.Context, cfg Config) error {
	return s.UpdateConfig(ctx, func(Config) (Config, error) { return cfg, nil })
}

func (s *Store) updateConfigTx(ctx context.Context, tx *sql.Tx, fn func(Config) (Config, error)) (int, error) {
	_, err := tx.StmtContext(ctx, s.lock).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	cfg, _, err := s.reloadTx(ctx, tx)
	if err != nil {
		return 0, err
	}

	newCfg, err := fn(*cfg)
	if err != nil {
		return 0, err
	}
	err = newCfg.Validate()
	if err != nil {
		return 0, err
	}

	data, err := jsonutil.Apply(cfg.data, newCfg)
	if err != nil {
		return 0, errors.Wrap(err, "merge config")
	}

	return s.SetConfigData(ctx, tx, data)
}

// Config will return the current config state.
func (s *Store) Config() Config {
	s.mx.RLock()
	cfg := s.rawCfg

	s.mx.RUnlock()
	return cfg
}
