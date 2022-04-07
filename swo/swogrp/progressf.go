package swogrp

import (
	"context"
	"fmt"
)

type ctxKey int

const (
	ctxKeyTask ctxKey = iota
)

type taskCtx struct {
	*Group
	TaskInfo
}

func withTask(ctx context.Context, grp *Group, info TaskInfo) context.Context {
	return context.WithValue(ctx, ctxKeyTask, &taskCtx{Group: grp, TaskInfo: info})
}

func task(ctx context.Context) *taskCtx {
	v := ctx.Value(ctxKeyTask)
	if v == nil {
		return nil
	}

	return v.(*taskCtx)
}

func Progressf(ctx context.Context, format string, args ...interface{}) {
	t := task(ctx)
	if t == nil {
		// not a running task
		return
	}

	t.TaskInfo.Status = fmt.Sprintf(format, args...)
	err := t.sendMessage(ctx, "task-progress", t.TaskInfo, false)
	if err != nil {
		t.Logger.Error(ctx, fmt.Errorf("send task-progress: %w", err))
	}
}
