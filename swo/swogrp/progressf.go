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
	return ctx.Value(ctxKeyTask).(*taskCtx)
}

func Progressf(ctx context.Context, format string, args ...interface{}) {
	t := task(ctx)
	t.TaskInfo.Status = fmt.Sprintf(format, args...)
	err := t.sendMessage(ctx, "task-progress", t.TaskInfo, false)
	if err != nil {
		t.Logger.Error(ctx, fmt.Errorf("send task-progress: %w", err))
	}
}
