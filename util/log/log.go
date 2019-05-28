package log

import (
	"context"
	"fmt"
	"os"

	"github.com/lib/pq"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

type logContextKey int

const (
	logContextKeyDebug logContextKey = iota
	logContextKeyRequestID
	logContextKeyFieldList
)

var defaultLogger = logrus.NewEntry(logrus.StandardLogger())
var defaultContext = context.Background()
var verbose = false
var stacks = false

// EnableStacks enables stack information via the Source field.
func EnableStacks() {
	stacks = true
}

// EnableJSON sets the output log format to JSON
func EnableJSON() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}
func init() {
	if terminal.IsTerminal(int(os.Stderr.Fd())) {
		logrus.SetFormatter(&terminalFormatter{})
	}
}

// EnableVerbose sets verbose logging. All debug messages will be logged.
func EnableVerbose() {
	verbose = true
}

func getLogger(ctx context.Context) *logrus.Entry {
	l := defaultLogger
	if ctx == nil {
		return l
	}

	l = l.WithFields(logrus.Fields(ContextFields(ctx)))

	rid := RequestID(ctx)
	if rid != "" {
		l = l.WithField("RequestID", rid)
	}

	return l
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func addSource(ctx context.Context, err error) context.Context {
	err = findRootSource(err) // always returns stackTracer
	if stacks {
		ctx = WithField(ctx, "Source", fmt.Sprintf("%+v", err.(stackTracer).StackTrace()))
	}
	if perr, ok := errors.Cause(err).(*pq.Error); ok && perr.Detail != "" {
		ctx = WithField(ctx, "SQLErrDetails", perr.Detail)
	}
	return ctx
}

type causer interface {
	Cause() error
}

func findRootSource(err error) error {
	var rootErr error
	for {
		if c, ok := err.(causer); ok {
			err = c.Cause()
		} else {
			break
		}
		if _, ok := err.(stackTracer); ok {
			rootErr = err
		}
	}
	if rootErr == nil {
		rootErr = errors.WithStack(err)
	}
	return rootErr
}

// Log will log an application error.
func Log(ctx context.Context, err error) {
	if err == nil {
		return
	}
	ctx = addSource(ctx, err)
	getLogger(ctx).WithError(err).Errorln()
}

// Logf will log application information.
func Logf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Printf(format, args...)
}

// Debugf will log the formatted string if the context has debug logging enabled.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		ctx = defaultContext
	}
	if !verbose {
		if v, _ := ctx.Value(logContextKeyDebug).(bool); !v {
			return
		}
	}

	getLogger(ctx).Infof(format, args...)
}

// Debug will log err if the context has debug logging enabled.
func Debug(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if ctx == nil {
		ctx = defaultContext
	}
	if !verbose {
		if v, _ := ctx.Value(logContextKeyDebug).(bool); !v {
			return
		}
	}
	ctx = addSource(ctx, err)

	getLogger(ctx).WithError(err).Infoln()
}

// EnableDebug will return a context where debug logging is enabled for it
// and all child contexts.
func EnableDebug(ctx context.Context) context.Context {
	return context.WithValue(ctx, logContextKeyDebug, true)
}
