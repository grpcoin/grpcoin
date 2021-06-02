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
	R     *redis.Client
	T     TimeProvider
	Trace trace.Tracer
}

func New(r *redis.Client, t TimeProvider, trace trace.Tracer) RateLimiter {
	return &rateLimiter{R: r, T: t, Trace: trace}
}

func (r *rateLimiter) Hit(ctx context.Context, id string, max int64) error {
	ctx, s := r.Trace.Start(ctx, "rate limiter")
	defer s.End()

	bucket := r.T().Truncate(time.Minute)
	k := RateKey(id, bucket)
	p := r.R.TxPipeline()
	incr := p.Incr(ctx, k)
	p.Expire(ctx, k, time.Minute*2)
	_, err := p.Exec(ctx)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to reach redis: %v", err))
	}
	cur := incr.Val()
	if cur >= max {
		return status.Error(codes.ResourceExhausted, fmt.Sprintf("rate limited: %d requests in the past minute (max: %d)", cur, max))
	}
	return nil
}

func RateKey(id string, t time.Time) string {
	return fmt.Sprintf("%s::%d", id, t.Unix())
}
