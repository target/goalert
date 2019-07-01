package switchover

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/version"
)

// Postgres channel names
const (
	StateChannel   = "goalert_switchover_state"
	ControlChannel = "goalert_switchover_control"
	DBIDChannel    = "goalert_switchover_db_id"
)

func (h *Handler) initNewDBListen(name string) error {
	u, err := url.Parse(name)
	if err != nil {
		return errors.Wrap(err, "parse db URL")
	}
	q := u.Query()
	q.Set("application_name", fmt.Sprintf("GoAlert %s (S/O Listener)", version.GitVersion()))
	u.RawQuery = q.Encode()
	name = u.String()

	l := pq.NewListener(name, 0, time.Second, h.listenEvent)

	err = l.Listen(DBIDChannel)
	if err != nil {
		l.Close()
		return err
	}
	err = l.Listen(StateChannel)
	if err != nil {
		l.Close()
		return err
	}

	go func() {
		for n := range l.NotificationChannel() {
			if n == nil {
				// nil can be sent, ignore
				continue
			}
			switch n.Channel {
			case DBIDChannel:
				h.mx.Lock()
				h.dbNextID = n.Extra
				h.mx.Unlock()
			case StateChannel:
				s, err := ParseStatus(n.Extra)
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
	u, err := url.Parse(name)
	if err != nil {
		return errors.Wrap(err, "parse db URL")
	}
	q := u.Query()
	q.Set("application_name", fmt.Sprintf("GoAlert %s (S/O Listener)", version.GitVersion()))
	u.RawQuery = q.Encode()
	name = u.String()

	h.l = pq.NewListener(name, 0, time.Second, h.listenEvent)

	err = h.l.Listen(StateChannel)
	if err != nil {
		h.l.Close()
		return err
	}
	err = h.l.Listen(ControlChannel)
	if err != nil {
		h.l.Close()
		return err
	}
	err = h.l.Listen(DBIDChannel)
	if err != nil {
		h.l.Close()
		return err
	}
	go h.listenLoop()
	return nil
}

func (h *Handler) pushState(s State) { h.stateCh <- s }

func (h *Handler) listenLoop() {
	ctx := context.Background()
	ctx = log.WithField(ctx, "NodeID", h.id)

	for n := range h.l.NotificationChannel() {
		if n == nil {
			// nil can be sent, ignore
			continue
		}
		switch n.Channel {
		case DBIDChannel:
			h.mx.Lock()
			h.dbID = n.Extra
			h.mx.Unlock()
		case StateChannel:
			s, err := ParseStatus(n.Extra)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "parse Status string"))
				continue
			}
			if s.State == StateAbort {
				go h.pushState(StateAbort)
			}
			h.statusCh <- s
		case ControlChannel:
			switch n.Extra {
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

			d, err := ParseDeadlineConfig(n.Extra, h.old.timeOffset)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "parse Deadlines string"))
				continue
			}
			h.controlCh <- d
		}
	}
}
func (h *Handler) listenEvent(ev pq.ListenerEventType, err error) {
	var event string
	switch ev {
	case pq.ListenerEventConnected:
		event = "connected"
	case pq.ListenerEventConnectionAttemptFailed:
		event = "connection attempt failed"
	case pq.ListenerEventDisconnected:
		event = "disconnected"
	case pq.ListenerEventReconnected:
		event = "reconnected"
	}
	if err != nil {
		log.Log(context.Background(), errors.Wrapf(err, "pq listen event '%s'", event))
	} else {
		log.Logf(context.Background(), "PQ Listen Event: %s", event)
	}
}
