package swo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/util/log"
)

type state struct {
	m *Manager

	stateName string

	status string

	nodes map[uuid.UUID]*Node

	taskClaimed bool

	taskID uuid.UUID
	cancel func()

	stateFn StateFunc

	mx sync.Mutex
}

func newState(ctx context.Context, m *Manager) (*state, error) {
	s := &state{
		m:         m,
		nodes:     make(map[uuid.UUID]*Node),
		stateFn:   StateUnknown,
		stateName: "unknown",
		cancel:    func() {},
	}

	return s, s.hello(ctx)
}

type StateFunc func(context.Context, *state, *swomsg.Message) StateFunc

func (s *state) Status() *Status {
	s.mx.Lock()
	defer s.mx.Unlock()

	var nodes []Node
	for _, n := range s.nodes {
		nodes = append(nodes, *n)
	}

	return &Status{
		Details: s.status,
		Nodes:   nodes,
	}
}

// IsIdle returns true before executing a switchover.
func (s Status) IsIdle() bool {
	for _, n := range s.Nodes {
		if n.Status != "idle" {
			return false
		}
	}
	return true
}

// IsDone returns true if the switchover has already been completed.
func (s Status) IsDone() bool {
	for _, n := range s.Nodes {
		if n.Status != "complete" {
			return false
		}
	}
	return true
}

// IsResetting returns true while the switchover is resetting.
func (s Status) IsResetting() bool {
	for _, n := range s.Nodes {
		if strings.HasPrefix(n.Status, "reset-") {
			return true
		}
	}
	return false
}

// IsExecuting returns true while the switchover is executing.
func (s Status) IsExecuting() bool {
	for _, n := range s.Nodes {
		if strings.HasPrefix(n.Status, "exec-") {
			return true
		}
	}
	return false
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

	// wait for poll interval before sending to new DB,
	// giving all nodes a chance to process
	time.Sleep(swomsg.PollInterval)
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
	if ok {
		n.NewValid = msg.Hello.IsNewDB
		return nil
	}

	s.nodes[msg.NodeID] = &Node{
		ID:       msg.NodeID,
		CanExec:  msg.Hello.CanExec,
		NewValid: msg.Hello.IsNewDB,
		Status:   msg.Hello.Status,
	}
	return nil
}

func (s *state) processFromOld(ctx context.Context, msg *swomsg.Message) error {
	s.update(msg)

	if msg.Reset != nil {
		s.cancel()
		s.nodes = make(map[uuid.UUID]*Node)
		s.m.app.Resume(ctx)
		s.taskID = msg.ID
		s.taskClaimed = false
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
		s.taskClaimed = false
		s.ackMessage(ctx, msg.ID)
		return StateExecWait
	}

	return StateIdle
}

func (s *state) isExecAck(msg *swomsg.Message) bool {
	if !s.m.canExec {
		return false
	}
	if msg.Ack == nil || !msg.Ack.Exec {
		return false
	}
	if msg.Ack.MsgID != s.taskID {
		return false
	}
	if s.taskClaimed {
		return false
	}

	if msg.NodeID != s.m.id {
		s.taskClaimed = true
		return false
	}

	return true
}

// StateExecWait is the state when the node is waiting for execution to be performed.
func StateExecWait(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "exec-wait"

	switch {
	case msg.Error != nil && msg.Error.MsgID == s.taskID:
		s.stateName = "error"
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil && msg.Done.MsgID == s.taskID:
		s.stateName = "idle"
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	case msg.Progress != nil:
		s.status = msg.Progress.Details
	case s.isExecAck(msg):
		s.StartTask(s.m.DoExecute)
		s.stateName = "exec-run"
		s.ackMessage(ctx, msg.ID)
		return StateResetRun
	}

	return StateExecWait
}

// StateExecRun is the state when the current node is executing the switchover.
func StateExecRun(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "exec-run"

	switch {
	case msg.Error != nil && msg.Error.MsgID == s.taskID:
		s.cancel()
		s.stateName = "error"
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil && msg.Done.MsgID == s.taskID:
		// already done, make sure we still cancel the context though
		s.cancel()
		s.stateName = "idle"
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	case msg.Progress != nil:
		s.status = msg.Progress.Details
	}

	return StateExecRun
}

// StateError is the state after a task failed.
func StateError(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "error"

	return StateError
}

// StateUnknown is the state after startup.
func StateUnknown(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "unknown"

	return StateError
}

// StateResetWait is the state when the node is waiting for a reset to be performed.
func StateResetWait(ctx context.Context, s *state, msg *swomsg.Message) StateFunc {
	s.stateName = "reset-wait"

	switch {
	case msg.Error != nil && msg.Error.MsgID == s.taskID:
		s.stateName = "error"
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil && msg.Done.MsgID == s.taskID:
		s.stateName = "idle"
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	case msg.Progress != nil:
		s.status = msg.Progress.Details
	case s.isExecAck(msg):
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
	case msg.Error != nil && msg.Error.MsgID == s.taskID:
		s.cancel()
		s.stateName = "error"
		s.ackMessage(ctx, msg.ID)
		return StateError
	case msg.Done != nil && msg.Done.MsgID == s.taskID:
		// already done, make sure we still cancel the context though
		s.cancel()
		s.stateName = "idle"
		s.ackMessage(ctx, msg.ID)
		return StateIdle
	case msg.Progress != nil:
		s.status = msg.Progress.Details
	}

	return StateResetRun
}
