package app

import (
	"context"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type recoverExporter struct {
	exp trace.Exporter
}

func (r recoverExporter) ExportSpan(s *trace.SpanData) {
	defer func() {
		err := recover()
		if err != nil {
			log.Log(context.Background(), errors.Errorf("export span (panic): %+v", err))
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
	defer func() {
		err := recover()
		if err != nil {
			log.Log(context.Background(), errors.Errorf("flush exporter (panic): %+v", err))
		}
	}()
	f.Flush()
}
