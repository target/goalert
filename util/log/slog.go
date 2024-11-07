package log

import (
	"context"
	"errors"
	"log/slog"
)

type slogHandler struct {
	l      *Logger
	parent *slogHandler
	attrs  []slog.Attr
	group  string
}

var _ slog.Handler = &slogHandler{}

func NewSlog(l *Logger) *slog.Logger {
	return slog.New(&slogHandler{l: l})
}

func (h *slogHandler) logger(ctx context.Context) *Logger {
	for {
		if h.l != nil {
			return h.l
		}
		if h.parent == nil {
			break
		}
		h = h.parent
	}

	return FromContext(ctx)
}

func (h *slogHandler) groupPrefix() string {
	var prefix string
	if h.parent != nil {
		prefix = h.parent.groupPrefix()
	}

	if h.group == "" {
		return prefix
	}

	return prefix + h.group + "."
}

func (h slogHandler) forEachAttr(ctx context.Context, f func(slog.Attr)) {
	if h.parent != nil {
		h.parent.forEachAttr(ctx, f)
	}

	prefix := h.groupPrefix()
	for _, a := range h.attrs {
		a.Key = prefix + a.Key
		f(a)
	}
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	l := h.logger(ctx)
	switch level {
	case slog.LevelDebug:
		return l.debug
	case slog.LevelInfo:
		return l.info
	}

	return level >= slog.LevelWarn
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	l := h.logger(ctx)

	e := l.l.WithContext(ctx)

	h.forEachAttr(ctx, func(a slog.Attr) {
		e = e.WithField(a.Key, a.Value)
	})

	switch record.Level {
	case slog.LevelDebug:
		e.Debug(record.Message)
	case slog.LevelInfo:
		e.Info(record.Message)
	case slog.LevelWarn:
		e.Warn(record.Message)
	case slog.LevelError:
		e.Error(record.Message)
	default:
		return errors.New("unknown log level")
	}
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{attrs: attrs, parent: h.parent}
}

func (h *slogHandler) WithGroup(group string) slog.Handler {
	return &slogHandler{group: group, parent: h.parent}
}
