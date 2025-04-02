package log

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type logContextKey int

const (
	logContextKeyDebug logContextKey = iota
	logContextKeyRequestID
	logContextKeyFieldList
	logContextKeyLogger
)

// Close is a convenience function for use with defer statements.
//
// Example: defer log.Close(ctx, fd.Close)
func Close(ctx context.Context, closeFn func() error) {
	if err := closeFn(); err != nil {
		Log(ctx, errors.Wrap(err, "close"))
	}
}

type Logger struct {
	debug  bool
	info   bool
	stacks bool
	l      *logrus.Logger

	errHooks []func(context.Context, error) context.Context
}

func (l *Logger) Logrus() *logrus.Logger { return l.l }

func NewLogger() *Logger {
	l := logrus.New()

	return &Logger{l: l, info: true}
}

func (l *Logger) BackgroundContext() context.Context {
	if l == nil {
		panic("nil logger")
	}

	return WithLogger(context.Background(), l)
}

func WithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, logContextKeyLogger, l)
}

func FromContext(ctx context.Context) *Logger {
	l, _ := ctx.Value(logContextKeyLogger).(*Logger)
	if l == nil {
		return NewLogger()
	}
	return l
}

// Write is a pass-through for the underlying logger's Write method.
func (l *Logger) Write(p []byte) (int, error) {
	return l.l.Writer().Write(p)
}

// SetOutput will change the log output.
func (l *Logger) SetOutput(out io.Writer) { l.l.SetOutput(out) }

// EnableStacks enables stack information via the Source field.
func (l *Logger) EnableStacks() { l.stacks = true }

// EnableJSON sets the output log format to JSON
func (l *Logger) EnableJSON() { l.l.SetFormatter(&logrus.JSONFormatter{}) }

// ErrorsOnly will disable all log output except errors.
func (l *Logger) ErrorsOnly() {
	l.debug = false
	l.info = false
}

// If EnableDebug is called, all debug messages will be logged.
func (l *Logger) EnableDebug() { l.debug = true }

func (l *Logger) entry(ctx context.Context) *logrus.Entry {
	e := logrus.NewEntry(l.l)
	if ctx == nil {
		return e
	}

	e = e.WithFields(logrus.Fields(ContextFields(ctx)))

	rid := RequestID(ctx)
	if rid != "" {
		e = e.WithField("RequestID", rid)
	}

	return e
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func (l *Logger) AddErrorMapper(mapper func(context.Context, error) context.Context) {
	l.errHooks = append(l.errHooks, mapper)
}

func (l *Logger) addSource(ctx context.Context, err error) context.Context {
	err = findRootSource(err) // always returns stackTracer
	if l.stacks {
		ctx = WithField(ctx, "Source", fmt.Sprintf("%+v", err.(stackTracer).StackTrace()))
	}
	for _, h := range l.errHooks {
		ctx = h(ctx, err)
	}
	return ctx
}

func findRootSource(err error) error {
	var rootErr error
	for {
		nextErr := errors.Unwrap(err)
		if nextErr == nil {
			break
		}
		err = nextErr

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
func Log(ctx context.Context, err error) { FromContext(ctx).Error(ctx, err) }

func (l *Logger) Error(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ctx = l.addSource(ctx, err)
	lg := l.entry(ctx).WithError(err)
	if errors.Is(err, context.Canceled) {
		lg.Debugln()
		return
	}

	lg.Errorln()
}

// Logf will log application information.
func Logf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Printf(ctx, format, args...)
}

func (l *Logger) Printf(ctx context.Context, format string, args ...interface{}) {
	if !l.info {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	l.entry(ctx).Printf(format, args...)
}

// Debugf will log the formatted string if the context has debug logging enabled.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).DebugPrintf(ctx, format, args...)
}

func (l *Logger) DebugPrintf(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	if v, _ := ctx.Value(logContextKeyDebug).(bool); !v && !l.debug {
		return
	}

	l.entry(ctx).Infof(format, args...)
}

// Debug will log err if the context has debug logging enabled.
func Debug(ctx context.Context, err error) { FromContext(ctx).DebugError(ctx, err) }

func (l *Logger) DebugError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if v, _ := ctx.Value(logContextKeyDebug).(bool); !v && !l.debug {
		return
	}
	ctx = l.addSource(ctx, err)

	l.entry(ctx).WithError(err).Infoln()
}

// WithDebug will enable debug logging for the context.
func WithDebug(ctx context.Context) context.Context {
	return context.WithValue(ctx, logContextKeyDebug, true)
}
