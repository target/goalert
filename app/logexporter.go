package app

import (
	"context"
	"github.com/target/goalert/util/log"

	"go.opencensus.io/trace"
)

type logExporter struct{}

func (l *logExporter) ExportSpan(span *trace.SpanData) {
	if !span.IsSampled() {
		return
	}
	ctx := log.WithField(context.Background(), "RequestID", span.TraceID.String())
	for _, a := range span.Annotations {
		log.Logf(log.WithFields(ctx, log.Fields(a.Attributes)), a.Message)
	}
}
