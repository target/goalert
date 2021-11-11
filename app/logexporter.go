package app

import (
	"github.com/target/goalert/util/log"

	"go.opencensus.io/trace"
)

type logExporter struct{ l *log.Logger }

func (l *logExporter) ExportSpan(span *trace.SpanData) {
	if !span.IsSampled() {
		return
	}
	ctx := log.WithField(l.l.BackgroundContext(), "RequestID", span.TraceID.String())
	for _, a := range span.Annotations {
		log.Logf(log.WithFields(ctx, log.Fields(a.Attributes)), a.Message)
	}
}
