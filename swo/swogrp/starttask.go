package swogrp

import (
	"context"
)

func (t *TaskMan) startTask(fn func(context.Context) error, successMsg string) {
	if t.cancelTask != nil {
		panic("already running a task")
	}

	ctx, cancel := context.WithCancel(t.cfg.Logger.BackgroundContext())
	t.cancelTask = cancel

	ackID := t.lastMsgID
	ctx = withMsgID(ctx, ackID)

	go func() {
		err := fn(ctx)
		if err != nil {
			t.mx.Lock()
			t.sendAck(ctx, "error", err.Error(), ackID)
			t.mx.Unlock()
			return
		}

		t.mx.Lock()
		t.sendAck(ctx, successMsg, nil, ackID)
		t.cancelTask()
		t.cancelTask = nil
		t.mx.Unlock()
	}()
}

func (t *TaskMan) cancel() {
	if t.cancelTask != nil {
		t.cancelTask()
	}
	t.cfg.Executor.Cancel()

	t.cancelTask = nil
	ctx := t.cfg.Logger.BackgroundContext()
	err := t.cfg.ResumeFunc(ctx)
	if err != nil {
		t.cfg.Logger.Error(ctx, err)
	}
}
