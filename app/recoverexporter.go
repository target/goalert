package app

import (
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type recoverExporter struct {
	logger *log.Logger
	exp    trace.Exporter
}

func (r recoverExporter) ExportSpan(s *trace.SpanData) {
	ctx := r.logger.Context()
	defer func() {
		err := recover()
		if err != nil {
			log.Log(ctx, errors.Errorf("export span (panic): %+v", err))
		}
	}()
	r.exp.ExportSpan(s)
}
func (r recoverExporter) Flush() {
	type flusher interface {
		Flush()
	}
	f, ok := r.exp.(flusher)
	if !ok {
		return
	}
	ctx := r.logger.Context()
	defer func() {
		err := recover()
		if err != nil {
			log.Log(ctx, errors.Errorf("flush exporter (panic): %+v", err))
		}
	}()
	f.Flush()
}
