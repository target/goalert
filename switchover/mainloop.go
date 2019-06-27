package switchover

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/util/log"
)

func (h *Handler) setState(ctx context.Context, newState State) {
	switch newState {
	case StatePaused, StatePauseWait, StatePausing:
	default:
		h.app.Resume()
	}
	if h.state == StateComplete && newState != StateStarting {
		return
	}
	if newState == StateAbort && !h.state.IsActive() {
		// already aborted
		return
	}
	if newState == h.state {
		return
	}

	h.mx.Lock()
	h.state = newState
	h.mx.Unlock()

	_, err := h.sendNotification.ExecContext(ctx, StateChannel, h.Status().serialize())
	if err != nil {
		log.Log(ctx, err)
	}
}
func (s State) oneOf(state []State) bool {
	for _, st := range state {
		if st == s {
			return true
		}
	}
	return false
}
func (h *Handler) allNodes(state ...State) bool {
	for _, s := range h.nodeStatus {
		if !s.State.oneOf(state) {
			return false
		}
	}
	return true
}

// CheckDBID will return true if the ID matches the current
// db-next ID. If there is no current ID, then it is set and returns true.
func (h *Handler) CheckDBID(id string) bool {
	h.mx.Lock()
	defer h.mx.Unlock()

	return h.dbID == id
}

// CheckDBNextID will return true if the ID matches the current
// db-next ID. If there is no current ID, then it is set and returns true.
func (h *Handler) CheckDBNextID(id string) bool {
	h.mx.Lock()
	defer h.mx.Unlock()

	return h.dbNextID == id
}
func (h *Handler) updateNodeStatus(ctx context.Context, s *Status) bool {
	if !h.CheckDBID(s.DBID) {
		log.Logf(ctx, "Switch-Over Abort: NodeID="+s.NodeID+" has mismatched DB ID")
		h.setState(ctx, StateAbort)
	}
	if !h.CheckDBNextID(s.DBNextID) {
		log.Logf(ctx, "Switch-Over Abort: NodeID="+s.NodeID+" has mismatched DB Next ID")
		h.setState(ctx, StateAbort)
	}

	oldStatus, ok := h.nodeStatus[s.NodeID]
	if oldStatus.State == s.State {
		return false
	}
	h.nodeStatus[s.NodeID] = *s
	if !ok {
		log.Logf(ctx, "Switch-Over Join: NodeID="+s.NodeID)
	}

	cnt := len(h.nodeStatus)

	statCount := make(map[State]int)
	for _, s := range h.nodeStatus {
		statCount[s.State]++
	}
	stats := make([]string, 0, len(statCount))
	for state, n := range statCount {
		stats = append(stats, fmt.Sprintf("%s %d/%d", state, n, cnt))
	}
	sort.Strings(stats)
	log.Logf(ctx, "Switch-Over State: %s", strings.Join(stats, ", "))
	if !ok && h.state != StateStarting && h.state != StateReady {
		h.setState(ctx, StateAbort)
	}
	return true
}

func (h *Handler) loop() {
	ctx := context.Background()
	ctx = log.WithField(ctx, "NodeID", h.id)
	statusUpdateT := time.NewTicker(3 * time.Second)
	defer statusUpdateT.Stop()
	var cfg DeadlineConfig
	deadline := time.NewTimer(0)
	deadline.Stop()
	var cancel func()
	pauseDone := make(chan struct{})

	abort := func() {
		if cancel != nil {
			cancel()
		}
		deadline.Stop()
		h.setState(ctx, StateAbort)
	}
	var lastDeadline string
	reset := func(name string, t time.Time) {
		deadline.Stop()
		deadline = time.NewTimer(time.Until(t))
		lastDeadline = name
	}

	for {
		select {
		case <-pauseDone:
			pauseDone = make(chan struct{})
			if h.state == StatePausing {
				reset("Switch-Over", cfg.AbsoluteDeadline())
				h.setState(ctx, StatePaused)
			}
		case <-deadline.C:
			switch h.state {
			case StateComplete:
				// already done
				continue
			case StateArmWait:
				// start the pause
				pauseDone = make(chan struct{})
				pauseCtx, pauseCancel := context.WithDeadline(ctx, cfg.PauseDeadline())
				pauseCtx = context.WithValue(pauseCtx, ctxValueDeadlines, cfg)
				cancel = pauseCancel
				reset("Pause", cfg.PauseDeadline())
				go func() {
					err := h.app.Pause(pauseCtx)
					if err != nil {
						log.Log(pauseCtx, errors.Wrap(err, "pause application"))
						abort()
						return
					}
					close(pauseDone)
				}()
				h.setState(ctx, StatePausing)
			default:
				log.Logf(ctx, "Switch-Over: Deadline reached (%s), aborting", lastDeadline)
				abort()
			}
			continue
		case d := <-h.controlCh:
			if h.state != StateReady {
				log.Logf(ctx, "Switch-Over: Control received but not ready, aborting")
				abort()
				continue
			}
			log.Logf(ctx, "Switch-Over: Control           BeginAt=%s", d.BeginAt.Format(time.RFC1123))
			log.Logf(ctx, "Switch-Over: Control ConsensusDeadline=%s", d.ConsensusDeadline().Format(time.RFC1123))
			log.Logf(ctx, "Switch-Over: Control           PauseAt=%s", d.PauseAt().Format(time.RFC1123))
			log.Logf(ctx, "Switch-Over: Control     PauseDeadline=%s", d.PauseDeadline().Format(time.RFC1123))
			log.Logf(ctx, "Switch-Over: Control  AbsoluteDeadline=%s", d.AbsoluteDeadline().Format(time.RFC1123))
			log.Logf(ctx, "Switch-Over: PAUSE BEGINS IN %s", time.Until(d.PauseAt()).String())

			cfg = *d
			reset("Consensus", cfg.ConsensusDeadline())
			h.setState(ctx, StateArmed)
		case <-statusUpdateT.C:
			if h.state == StateStarting && h.app.Status() == lifecycle.StatusReady {
				h.setState(ctx, StateReady)
				continue
			}

			_, err := h.sendNotification.ExecContext(ctx, StateChannel, h.Status().serialize())
			if err != nil {
				log.Log(ctx, err)
			}
		case state := <-h.stateCh:
			switch state {
			case StateAbort:
				log.Logf(ctx, "Switch-Over: Got abort signal.")
				abort()
				continue
			case StateStarting:
				log.Logf(ctx, "Switch-Over: Got reset signal.")
				abort() //reset
				h.nodeStatus = make(map[string]Status)
			default:
				log.Logf(ctx, "Switch-Over: Got signal '%s'.", state)
			}

			h.setState(ctx, state)
		case stat := <-h.statusCh:
			if !h.updateNodeStatus(ctx, stat) {
				continue
			}
			switch h.state {
			case StateArmed:
				if h.allNodes(StateArmed, StateArmWait) {
					log.Logf(ctx, "Switch-Over: Consensus reached after %s", time.Since(cfg.BeginAt).String())
					reset("PauseAt", cfg.PauseAt())
					h.setState(ctx, StateArmWait)
				}
			case StatePaused:
				if h.allNodes(StatePaused, StatePauseWait) {
					log.Logf(ctx, "Switch-Over: World paused after %s", time.Since(cfg.PauseAt()).String())
					h.setState(ctx, StatePauseWait)
				}
			}
		}
	}
}
