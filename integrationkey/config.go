package integrationkey

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
)

type dbConfig struct {
	Version int // should be 1
	V1      Config
}

type Config struct {
	StopAfterFirstMatchingRule bool

	// Rules is a list of rules to apply to the alert.
	Rules []Rule

	Suppression []SuppWindow

	DefaultActions []Action
}

type Rule struct {
	ID uuid.UUID

	// Name is the name of the rule.
	Name string

	// Description is a description of the rule.
	Description string

	ConditionExpr string

	DedupConfig DedupConfig

	Actions []Action
}

type DedupConfig struct {
	IDExpr        string
	WindowSeconds int
}

type SuppWindow struct {
	Start time.Time
	End   time.Time

	FilterExpr string
}

type Action struct {
	// Type is the type of action to perform, like slack, email, or alert.
	Type string

	// StaticParams are parameters that are always the same for this action (e.g., the channel ID).
	StaticParams map[string]string

	// DynamicParams are parameters that are determined at runtime (e.g., the message to send).
	// The keys are the parameter names, and the values are the Expr expression strings.
	DynamicParams map[string]string
}

func (s *Store) Config(ctx context.Context, db gadb.DBTX, keyID uuid.UUID) (*Config, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	data, err := gadb.New(db).IntKeyGetConfig(ctx, keyID)
	if errors.Is(err, sql.ErrNoRows) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg dbConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version: %d", cfg.Version)
	}

	return &cfg.V1, nil
}

func (s *Store) SetConfig(ctx context.Context, db gadb.DBTX, keyID uuid.UUID, cfg *Config) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	if cfg == nil {
		return gadb.New(db).IntKeyDeleteConfig(ctx, keyID)
	}

	// ensure all rule IDs are set
	for i := range cfg.Rules {
		if cfg.Rules[i].ID == uuid.Nil {
			cfg.Rules[i].ID = uuid.New()
		}
	}

	data, err := json.Marshal(dbConfig{Version: 1, V1: *cfg})
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = gadb.New(db).IntKeySetConfig(ctx, gadb.IntKeySetConfigParams{
		ID:     keyID,
		Config: data,
	})
	if err != nil {
		return err
	}

	return nil
}
