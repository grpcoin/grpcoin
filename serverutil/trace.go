// Copyright 2021 Ahmet Alp Balkan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serverutil

import (
	"context"
	"log"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func GetTracer(component string, onCloud bool) (tracer oteltrace.Tracer, flush func(logger *zap.Logger)) {
	var traceExporter trace.SpanExporter
	if !onCloud {
		return oteltrace.NewNoopTracerProvider().Tracer(""), func(logger *zap.Logger) {
		}
	}
	gcp, err := texporter.NewExporter(
		texporter.WithTraceClientOptions([]option.ClientOption{
			// option.WithTelemetryDisabled(),
		})) // don't trace the trace client itself
	if err != nil {
		log.Fatal("failed to initialize gcp trace exporter", zap.Error(err))
	}
	traceExporter = gcp
	tr := trace.NewTracerProvider(trace.WithBatcher(traceExporter,
		trace.WithMaxQueueSize(5000),
		trace.WithMaxExportBatchSize(1000),
	), trace.WithSampler(trace.TraceIDRatioBased(0.1)))

	otel.SetTracerProvider(tr)
	tp := otel.GetTracerProvider().Tracer(component)

	octrace.DefaultTracer = opencensus.NewTracer(tp)
	return tp, func(log *zap.Logger) {
		log.Debug("force flushing trace spans")
		// we don't use the main ctx here as it'll be cancelled by the time this is executed
		if err := tr.ForceFlush(context.TODO()); err != nil {
			log.Warn("failed to flush tracer", zap.Error(err))
		}
	}
}
