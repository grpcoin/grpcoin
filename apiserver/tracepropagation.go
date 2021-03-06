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

package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpHeader = "X-Cloud-Trace-Context"
)

// HTTPFormat propagator serializes SpanContext to/from HTTP Headers.
type HTTPFormat struct{}

var _ propagation.TextMapPropagator = &HTTPFormat{}

// Inject injects a context into the carrier as HTTP headers.
func (hf HTTPFormat) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()

	if !sc.TraceID().IsValid() || !sc.SpanID().IsValid() {
		return
	}

	spanID := sc.SpanID()
	sid := binary.BigEndian.Uint64(spanID[:])
	header := fmt.Sprintf("%s/%d;o=%d", sc.TraceID().String(), sid, sc.TraceFlags())
	carrier.Set(httpHeader, header)
}

// Extract extracts a context from the carrier if it contains HTTP headers.
func (hf HTTPFormat) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {

	if h := carrier.Get(httpHeader); h != "" {
		sc, err := extract(h)
		if err == nil && sc.IsValid() {
			return trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}

	return ctx
}

func extract(h string) (trace.SpanContext, error) {
	sc := trace.SpanContext{}

	// Parse the trace id field.
	slash := strings.Index(h, `/`)
	if slash == -1 {
		return sc, errors.New("failed to parse value")
	}
	tid, h := h[:slash], h[slash+1:]

	traceID, err := trace.TraceIDFromHex(tid)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}

	sc = sc.WithTraceID(traceID)

	// Parse the span id field.
	spanstr := h
	semicolon := strings.Index(h, `;`)
	if semicolon != -1 {
		spanstr, h = h[:semicolon], h[semicolon+1:]
	}
	sid, err := strconv.ParseUint(spanstr, 10, 64)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}
	spanID := sc.SpanID()
	binary.BigEndian.PutUint64(spanID[:], sid)
	sc = sc.WithSpanID(spanID)

	// Parse the options field, options field is optional.
	if !strings.HasPrefix(h, "o=") {
		return sc, errors.New("failed to parse value")
	}

	var sampled trace.TraceFlags
	if h[2:] == "1" {
		sampled = trace.FlagsSampled
	}
	sc = sc.WithTraceFlags(sampled)

	return sc, nil
}

// Fields is propagation keys
func (hf HTTPFormat) Fields() []string {
	return []string{httpHeader}
}
