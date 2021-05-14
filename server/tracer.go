package main

import (
	"context"

	"go.opentelemetry.io/otel/sdk/trace"
)

var _ trace.SpanExporter = dummyTraceExporter{}

type dummyTraceExporter struct{}

func (d dummyTraceExporter) ExportSpans(ctx context.Context, ss []*trace.SpanSnapshot) error {
	return nil
}

func (d dummyTraceExporter) Shutdown(ctx context.Context) error {
	return nil
}
