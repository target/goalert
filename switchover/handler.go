package switchover

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"

	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Handler struct {
	old, new  *dbState
	id        string
	dbNextURL string

	dbID, dbNextID string

	sendNotification *sql.Stmt
	nodeStatus       map[string]Status
	l                *sqlutil.Listener

	statusCh  chan *Status
	controlCh chan *DeadlineConfig
	stateCh   chan State

	mx sync.Mutex

	state State
	app   App
}

type App interface {
	Pause(context.Context) error
	Resume()
	Status() lifecycle.Status
}

func NewHandler(ctx context.Context, oldC, newC driver.Connector, oldURL, newURL string) (*Handler, error) {
	h := &Handler{
		id:         uuid.NewV4().String(),
		stateCh:    make(chan State),
		statusCh:   make(chan *Status),
		controlCh:  make(chan *DeadlineConfig),
		nodeStatus: make(map[string]Status),
		state:      StateStarting,
		dbNextURL:  newURL,
	}
	var err error
	h.old, err = newDBState(ctx, oldC)
	if err != nil {
		return nil, errors.Wrap(err, "init old db")
	}
	log.Logf(ctx, "DB_URL time offset "+h.old.timeOffset.String())
	h.sendNotification, err = h.old.db.PrepareContext(ctx, `select pg_notify($1, $2)`)
	if err != nil {
		return nil, errors.Wrap(err, "prepare notify statement")
	}

	h.new, err = newDBState(ctx, newC)
	if err != nil {
		return nil, errors.Wrap(err, "init new db")
	}
	log.Logf(ctx, "DB_URL_NEXT time offset "+h.new.timeOffset.String())
	diff := h.new.timeOffset - h.old.timeOffset
	if diff < 0 {
		diff = -diff
	}
	log.Logf(ctx, "DB time offsets differ by "+diff.String())

	err = h.initListen(oldURL)
	if err != nil {
		return nil, errors.Wrap(err, "init DB listener")
	}
	err = h.initNewDBListen(newURL)
	if err != nil {
		return nil, errors.Wrap(err, "init DB-next listener")
	}

	go h.loop()
	return h, nil
}

func (h *Handler) Abort() {
	h.stateCh <- StateAbort
}

func (h *Handler) Status() *Status {
	h.mx.Lock()
	defer h.mx.Unlock()
	return h.status()
}
func (h *Handler) status() *Status {
	return &Status{
		State:    h.state,
		NodeID:   h.id,
		Offset:   h.old.timeOffset,
		DBID:     h.dbID,
		DBNextID: h.dbNextID,
	}
}

func (h *Handler) DB() *sql.DB {
	return sql.OpenDB(h)
}

func (h *Handler) Connect(ctx context.Context) (c driver.Conn, err error) {
	c, err = h.old.dbc.Connect(ctx)
	if err != nil {
		return nil, err
	}
	_, err = c.(driver.ExecerContext).ExecContext(ctx, `select pg_advisory_lock_shared($1)`, []driver.NamedValue{{Ordinal: 1, Value: int64(lock.GlobalSwitchOver)}})
	if err != nil {
		c.Close()
		return nil, err
	}

	rows, err := c.(driver.QueryerContext).
		QueryContext(ctx, `select current_state from switchover_state`,
			nil,
		)
	if err != nil {
		c.Close()
		return nil, err
	}

	scan := make([]driver.Value, 1)
	err = rows.Next(scan)
	if err != nil {
		c.Close()
		return nil, err
	}
	var state string
	switch t := scan[0].(type) {
	case string:
		state = t
	case []byte:
		state = string(t)
	default:
		return nil, fmt.Errorf("expected string for current_state value, got %t", t)
	}

	if rows.Next(nil) != io.EOF {
		c.Close()
		return nil, errors.New("expected single row in switchover_state table")
	}
	rows.Close()

	switch state {
	case "idle", "in_progress":
		return c, nil
	case "use_next_db":
		c.Close()
		h.stateCh <- StateComplete
		return h.new.dbc.Connect(ctx)
	}

	return nil, fmt.Errorf("invalid state %s", state)
}

func (h *Handler) Driver() driver.Driver { return nil }

func (h *Handler) SetApp(app App) {
	h.app = app
}
