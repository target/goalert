package swogrp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/swo/swomsg"
)

type TaskMgr struct {
	local Node

	cfg     Config
	nodes   map[uuid.UUID]Node
	paused  map[uuid.UUID]struct{}
	waitMsg map[uuid.UUID]chan struct{}
	state   ClusterState

	cancelTask func()
	lastMsgID  uuid.UUID
	leaderID   uuid.UUID

	lastError     string
	lastStatus    string
	pendingStatus string
	mx            sync.Mutex
}

func NewTaskMgr(ctx context.Context, cfg Config) (*TaskMgr, error) {
	t := &TaskMgr{
		cfg: cfg,
		local: Node{
			ID:    uuid.New(),
			OldID: cfg.OldID,
			NewID: cfg.NewID,

			CanExec: cfg.CanExec,
		},
		nodes:   make(map[uuid.UUID]Node),
		paused:  make(map[uuid.UUID]struct{}),
		waitMsg: make(map[uuid.UUID]chan struct{}),
	}

	t.sendAck(ctx, "hello", t.local, uuid.Nil)

	return t, nil
}

func (t *TaskMgr) Init() {
	go t.statusLoop()
	go t.messageLoop()
}

func (t *TaskMgr) allNodesPaused() bool {
	for _, n := range t.nodes {
		_, ok := t.paused[n.ID]
		if !ok {
			return false
		}
	}

	return true
}

func (t *TaskMgr) statusLoop() {
	ctx := t.cfg.Logger.BackgroundContext()
	// debounce/throttle status messages
	var lastStatus string
	for range time.NewTicker(time.Second).C {
		t.mx.Lock()
		status := t.pendingStatus
		id := t.lastMsgID
		t.mx.Unlock()
		if status == lastStatus || status == "" {
			continue
		}

		lastStatus = status
		t.sendAck(ctx, "status", status, id)
	}
}

func (t *TaskMgr) messageLoop() {
	ctx := t.cfg.Logger.BackgroundContext()
	for msg := range t.cfg.Messages.Events() {
		t.mx.Lock()
		if ch, ok := t.waitMsg[msg.ID]; ok {
			close(ch)
			delete(t.waitMsg, msg.ID)
		}
		switch {
		case msg.Type == "reset":
			t.state = ClusterStateResetting
			t.cancel()
			t.leaderID = uuid.Nil
			t.lastStatus = ""
			t.lastError = ""
			t.pendingStatus = ""
			t.lastMsgID = msg.ID
			for id := range t.nodes {
				delete(t.nodes, id)
				delete(t.paused, id)
			}
			t.sendAck(ctx, "hello", t.local, t.lastMsgID)
		case t.state == ClusterStateResetting && msg.Type == "hello" && msg.AckID == t.lastMsgID:
			var n Node
			err := json.Unmarshal(msg.Data, &n)
			if err != nil {
				t.sendAck(ctx, "error", fmt.Sprintf("unmarshal hello: %v", err), msg.ID)
				continue
			}
			t.nodes[msg.Node] = n
			if t.leaderID != uuid.Nil {
				// already have leader
				break
			}
			if !n.CanExec {
				// can't be leader
				break
			}
			t.leaderID = n.ID
			if t.leaderID != t.local.ID {
				// not us
				break
			}

			// leader, start timer
			t.startTask(resetDelay, "reset-end")
		case t.state == ClusterStateResetting && msg.Type == "reset-end" && msg.AckID == t.lastMsgID:
			t.lastMsgID = msg.ID
			t.state = ClusterStateIdle
		case t.state == ClusterStateIdle && msg.Type == "execute" && msg.AckID == t.lastMsgID:
			t.state = ClusterStateSyncing
			t.lastMsgID = msg.ID
			if t.leaderID != t.local.ID {
				break
			}

			t.startTask(t.cfg.Executor.Sync, "pause")
		case t.state == ClusterStateSyncing && msg.Type == "pause" && msg.AckID == t.lastMsgID:
			t.state = ClusterStatePausing
			t.lastMsgID = msg.ID

			t.startTask(func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
				return t.cfg.PauseFunc(ctx)
			}, "paused")
		case t.state == ClusterStatePausing && msg.Type == "paused" && msg.AckID == t.lastMsgID:
			t.paused[msg.Node] = struct{}{}
			if !t.allNodesPaused() {
				break
			}
			t.state = ClusterStateExecuting
			t.lastMsgID = msg.ID
			if t.leaderID != t.local.ID {
				break
			}
			t.startTask(t.cfg.Executor.Exec, "done")
		case t.state == ClusterStateExecuting && msg.Type == "done" && msg.AckID == t.lastMsgID:
			t.state = ClusterStateDone
		case msg.Type == "status":
			if msg.AckID != t.lastMsgID {
				break
			}
			t.lastStatus = t.parseString(msg.Data)
		case msg.Type == "error":
			if msg.AckID != t.lastMsgID {
				break
			}
			t.lastError = t.parseString(msg.Data)
			fallthrough
		case msg.Type == "cancel":
			t.cancel()
			t.state = ClusterStateUnknown
		default:
			if t.state != ClusterStateUnknown {
				// only report on change
				t.sendAck(ctx, "cancel", "unexpected or invalid message", msg.ID)
			}
			t.cancel()
			t.state = ClusterStateUnknown
		}
		t.mx.Unlock()
	}
}

func (t *TaskMgr) parseString(data json.RawMessage) string {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		t.cfg.Logger.Error(context.Background(), fmt.Errorf("unmarshal string: %w", err))
		return ""
	}
	return s
}

func resetDelay(ctx context.Context) error {
	t := time.NewTimer(3 * time.Second)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (t *TaskMgr) sendAckWait(ctx context.Context, msgType string, v interface{}, ackID uuid.UUID) <-chan struct{} {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("marshal %s: %w", msgType, err))
	}

	ch := make(chan struct{})
	id := uuid.New()
	t.mx.Lock()
	t.waitMsg[id] = ch
	t.mx.Unlock()

	err = t.cfg.Messages.Append(ctx, swomsg.Message{
		ID:    id,
		Node:  t.local.ID,
		AckID: ackID,

		Type: msgType,
		Data: data,
	})
	if err != nil {
		close(ch)
		t.cfg.Logger.Error(ctx, fmt.Errorf("append %s: %w", msgType, err))
	}

	return ch
}

func (t *TaskMgr) sendAck(ctx context.Context, msgType string, v interface{}, ackID uuid.UUID) {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("marshal %s: %w", msgType, err))
	}

	err = t.cfg.Messages.Append(ctx, swomsg.Message{
		ID:    uuid.New(),
		Node:  t.local.ID,
		AckID: ackID,

		Type: msgType,
		Data: data,
	})
	if err != nil {
		t.cfg.Logger.Error(ctx, fmt.Errorf("append %s: %w", msgType, err))
	}
}

type taskCtx string

func withMsgID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, taskCtx("msgID"), id)
}
func msgID(ctx context.Context) uuid.UUID { return ctx.Value(taskCtx("msgID")).(uuid.UUID) }

func (t *TaskMgr) Statusf(ctx context.Context, format string, args ...interface{}) {
	t.mx.Lock()
	if t.lastMsgID == msgID(ctx) {
		t.pendingStatus = fmt.Sprintf(format, args...)
	}
	t.mx.Unlock()
}

func (t *TaskMgr) Cancel(ctx context.Context) error {
	ch := t.sendAckWait(ctx, "cancel", nil, uuid.Nil)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
	}

	return nil
}

func (t *TaskMgr) Reset(ctx context.Context) error {
	ch := t.sendAckWait(ctx, "reset", nil, uuid.Nil)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
	}

	return nil
}

func (t *TaskMgr) Execute(ctx context.Context) error {
	t.mx.Lock()
	state := t.state
	t.mx.Unlock()
	if state != ClusterStateIdle {
		return fmt.Errorf("cannot execute unless idle")
	}

	ch := t.sendAckWait(ctx, "execute", nil, t.lastMsgID)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
	}

	return nil
}
