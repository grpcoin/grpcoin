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

package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TimeProvider func() time.Time

type RateLimiter interface {
	Hit(ctx context.Context, id string, max int64) error
}

type rateLimiter struct {
	R      *redis.Client
	Trace  trace.Tracer
	T      TimeProvider
	Window time.Duration
}

func New(r *redis.Client, t TimeProvider, trace trace.Tracer, windowSize time.Duration) RateLimiter {
	return &rateLimiter{R: r, T: t, Trace: trace, Window: windowSize}
}

func (r *rateLimiter) Hit(ctx context.Context, key string, max int64) error {
	ctx, s := r.Trace.Start(ctx, "rate limiter")
	defer s.End()

	bucket := r.T().Truncate(r.Window)
	k := RateKey(key, bucket)
	p := r.R.TxPipeline()
	incr := p.Incr(ctx, k)
	p.Expire(ctx, k, r.Window*2) // no need for *2 here, but keep it for debugging
	_, err := p.Exec(ctx)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to reach redis: %v", err))
	}
	cur := incr.Val()
	if cur > max {
		return status.Error(codes.ResourceExhausted, fmt.Sprintf("rate limited: %d requests in the past %v (max: %d)", cur, r.Window, max))
	}
	return nil
}

func RateKey(id string, t time.Time) string {
	return fmt.Sprintf("rate::%s::%d", id, t.Unix())
}
