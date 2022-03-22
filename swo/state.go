package swo

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/util/log"
)

type state struct {
	m *Manager

	stateName string

	status string

	nodes map[uuid.UUID]*Node

	taskID uuid.UUID
	cancel func()

	stateFn StateFunc

	mx sync.Mutex
}

func newState(ctx context.Context, m *Manager) (*state, error) {
	s := &state{
		m:         m,
		nodes:     make(map[uuid.UUID]*Node),
		stateFn:   StateIdle,
		stateName: "unknown",
		cancel:    func() {},
	}

	return s, s.hello(ctx)
}

type StateFunc func(context.Context, *state, *swomsg.Message) StateFunc

func (s *state) Status() *Status {
	s.mx.Lock()
	defer s.mx.Unlock()

	isIdle := true
	isDone := true
	var nodes []Node
	for _, n := range s.nodes {
		nodes = append(nodes, *n)
		isIdle = isIdle && n.Status == "idle"
		isDone = isDone && n.Status == "complete"
	}

	return &Status{
		Details: s.status,
		Nodes:   nodes,

		IsDone: isDone,
		IsIdle: isIdle,
	}
}

func (s *state) ackMessage(ctx context.Context, msgID uuid.UUID) {
	err := s.m.msgLog.Append(ctx, swomsg.Ack{MsgID: msgID, Status: s.stateName, Exec: s.m.canExec})
	if err != nil {
		log.Log(ctx, err)
	}
}

func (s *state) update(msg *swomsg.Message) {
	s.mx.Lock()
	defer s.mx.Unlock()

	n, ok := s.nodes[msg.NodeID]
	if !ok {
		n = &Node{
			ID: msg.NodeID,
		}
		s.nodes[msg.NodeID] = n
	}

	switch {
	case msg.Hello != nil:
		n.OldValid = msg.Hello.IsOldDB
		n.Status = msg.Hello.Status
		n.CanExec = msg.Hello.CanExec
	case msg.Ack != nil:
		n.Status = msg.Ack.Status
		n.CanExec = msg.Ack.Exec
	case msg.Progress != nil:
		s.status = msg.Progress.Details
	case msg.Error != nil:
		s.status = "error: " + msg.Error.Details
	case msg.Done != nil:
		s.status = ""
	}
}

func (s *state) taskDone(ctx context.Context, err error) {
	if err != nil {
		err = s.m.msgLog.Append(ctx, swomsg.Error{MsgID: s.taskID, Details: err.Error()})
	} else {
		err = s.m.msgLog.Append(ctx, swomsg.Done{MsgID: s.taskID})
	}
	if err != nil {
		log.Log(ctx, err)
	}
}

func (s *state) hello(ctx context.Context) error {
	err := s.m.msgLog.Append(ctx, swomsg.Hello{IsOldDB: true, Status: s.stateName, CanExec: s.m.canExec})
	if err != nil {
		return err
	}
	err = s.m.nextMsgLog.Append(ctx, swomsg.Hello{IsNewDB: true, Status: s.stateName, CanExec: s.m.canExec})
	if err != nil {
		return err
	}
	return nil
}

func (s *state) processFromNew(ctx context.Context, msg *swomsg.Message) error {
	if msg.Hello == nil {
		return fmt.Errorf("unexpected message to NEW DB: %v", msg)
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	n, ok := s.nodes[msg.NodeID]
	if !ok {
		n = &Node{
			ID: msg.NodeID,
		}
		s.nodes[msg.NodeID] = n
	}
	n.NewValid = msg.Hello.IsNewDB
	return nil
}

func (s *state) processFromOld(ctx context.Context, msg *swomsg.Message) error {
	s.update(msg)

	if msg.Reset != nil {
		s.cancel()
		s.nodes = make(map[uuid.UUID]*Node)
		s.m.app.Resume(ctx)
		s.taskID = msg.ID
		s.stateName = "reset-wait"
		s.stateFn = StateResetWait
		err := s.hello(ctx)
		if err != nil {
			return err
		}
		if s.m.canExec {
			s.ackMessage(ctx, msg.ID)
		}
		return nil
	}

	s.stateFn = s.stateFn(ctx, s, msg)
	if msg.Ping != nil {
		s.ackMessage(ctx, msg.ID)
	}

	return nil
}

func (s *state) StartTask(task func(context.Context) error) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go func() { s.taskDone(ctx, task(ctx)) }()
}

// StateIdle is the state when the node is idle.
func StateIdle(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "idle"

	switch {
	case msg.Execute != nil:
		s.taskID = msg.ID
		s.stateName = "exec-wait"
		s.ackMessage(ctx, msg.ID)
		return StateExecWait
	case msg.Plan != nil:
	}

	return StateIdle
}

// StateExecWait is the state when the node is waiting for execution to be performed.
func StateExecWait(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "exec-wait"

	switch {
	case msg.Error != nil:
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil:
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	case msg.Ack != nil && s.m.canExec:
		if msg.Ack.MsgID != s.taskID {
			// ack for a different message
			break
		}
		if msg.NodeID != s.m.id {
			// claimed by another node
			s.taskID = uuid.Nil
			break
		}
		s.StartTask(s.m.DoExecute)
		s.stateName = "exec-run"
		s.ackMessage(ctx, msg.ID)
		return StateResetRun
	}

	return StateExecWait
}

// StateExecRun is the state when the current node is executing the switch-over.
func StateExecRun(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "exec-run"

	switch {
	case msg.Error != nil:
		s.cancel()
		s.stateName = "error"
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil:
		// already done, make sure we still cancel the context though
		s.cancel()
		s.stateName = "idle"
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	}

	return StateExecRun
}

// StateError is the state after a task failed.
func StateError(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "error"

	return StateError
}

// StateResetWait is the state when the node is waiting for a reset to be performed.
func StateResetWait(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "reset-wait"

	switch {
	case msg.Error != nil:
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil:
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	case msg.Ack != nil && s.m.canExec:
		if msg.Ack.MsgID != s.taskID {
			// ack for a different message
			break
		}
		if msg.NodeID != s.m.id {
			// claimed by another node
			s.taskID = uuid.Nil
			break
		}
		s.StartTask(s.m.DoReset)
		s.stateName = "reset-run"
		s.ackMessage(ctx, msg.ID)
		return StateResetRun
	}

	return StateResetWait
}

// StateResetRun is the state when the current node is performing a reset.
func StateResetRun(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "reset-run"

	switch {
	case msg.Error != nil:
		s.cancel()
		s.stateName = "error"
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil:
		// already done, make sure we still cancel the context though
		s.cancel()
		s.stateName = "idle"
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	}

	return StateResetRun
}
