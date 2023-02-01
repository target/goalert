package swogrp

import (
	"context"

	"github.com/google/uuid"
	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/util/log"
)

type TaskFn func(context.Context) error

// Config is the configuration for a switchover group.
type Config struct {
	// NodeID is the unique ID of the current node.
	NodeID uuid.UUID

	// CanExec indicates this member is allowed to execute tasks.
	CanExec bool

	// OldID and NewID represents the database IDs of the old and new databases, respectively.
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
