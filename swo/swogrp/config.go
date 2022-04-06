package swogrp

import (
	"context"

	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/util/log"
)

type Config struct {
	CanExec bool

	Logger  *log.Logger
	MainLog *swomsg.Log
	NextLog *swomsg.Log

	ResetFunc   func(context.Context) error
	ExecuteFunc func(context.Context) error
	PauseFunc   func(context.Context) error
	ResumeFunc  func(context.Context) error
}
