package switchover

import (
	"context"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/version"
)

// Postgres channel names
const (
	StateChannel   = "goalert_switchover_state"
	ControlChannel = "goalert_switchover_control"
	DBIDChannel    = "goalert_switchover_db_id"
)

func (h *Handler) initNewDBListen(name string) error {
	cfg, err := pgx.ParseConnectionString(name)
	if err != nil {
		return err
	}
	cfg.RuntimeParams["application_name"] = fmt.Sprintf("GoAlert %s (S/O Listener)", version.GitVersion())
	l, err := sqlutil.NewListener(context.Background(), sqlutil.ConfigConnector(cfg), DBIDChannel, StateChannel)
	if err != nil {
		return err
	}

	go func() {
		defer l.Close()
		for n := range l.Notifications() {
			switch n.Channel {
			case DBIDChannel:
				h.mx.Lock()
				h.dbNextID = n.Payload
				h.mx.Unlock()
			case StateChannel:
				s, err := ParseStatus(n.Payload)
				if err != nil {
					log.Log(context.Background(), errors.Wrap(err, "parse Status string"))
					continue
				}
				if s.State == StateAbort {
					go h.pushState(StateAbort)
				}
				h.statusCh <- s
			}
		}
	}()

	return nil
}
func (h *Handler) initListen(name string) error {
	cfg, err := pgx.ParseConnectionString(name)
	if err != nil {
		return err
	}
	cfg.RuntimeParams["application_name"] = fmt.Sprintf("GoAlert %s (S/O Listener)", version.GitVersion())

	h.l, err = sqlutil.NewListener(context.Background(), sqlutil.ConfigConnector(cfg), DBIDChannel, StateChannel, ControlChannel)
	if err != nil {
		return err
	}

	go h.listenLoop()
	return nil
}

func (h *Handler) pushState(s State) { h.stateCh <- s }

func (h *Handler) listenLoop() {
	ctx := context.Background()
	ctx = log.WithField(ctx, "NodeID", h.id)
	defer h.l.Close()

	for n := range h.l.Notifications() {
		if n == nil {
			// nil can be sent, ignore
			continue
		}
		switch n.Channel {
		case DBIDChannel:
			h.mx.Lock()
			h.dbID = n.Payload
			h.mx.Unlock()
		case StateChannel:
			s, err := ParseStatus(n.Payload)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "parse Status string"))
				continue
			}
			if s.State == StateAbort {
				go h.pushState(StateAbort)
			}
			h.statusCh <- s
		case ControlChannel:
			switch n.Payload {
			case "done":
				go h.pushState(StateComplete)
				continue
			case "abort":
				go h.pushState(StateAbort)
				continue
			case "reset":
				go h.pushState(StateStarting)
				continue
			}

			d, err := ParseDeadlineConfig(n.Payload, h.old.timeOffset)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "parse Deadlines string"))
				continue
			}
			h.controlCh <- d
		}
	}
}
