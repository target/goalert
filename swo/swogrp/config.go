package swogrp

import (
	"context"

	"github.com/google/uuid"
	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/util/log"
)

type TaskFn func(context.Context) error

type Config struct {
	NodeID uuid.UUID

	CanExec      bool
	OldID, NewID uuid.UUID

	Logger   *log.Logger
	Messages *swomsg.Log

	PauseFunc  TaskFn
	ResumeFunc TaskFn

	Executor Executor
}

type Executor interface {
	Sync(context.Context) error
	Exec(context.Context) error

	Cancel()
}
