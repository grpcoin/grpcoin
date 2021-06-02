package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRateLimiter_Hit(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer s.Close()

	var tt time.Time
	rl := rateLimiter{
		R:     rc,
		T:     func() time.Time { return tt },
		Trace: trace.NewNoopTracerProvider().Tracer("")}

	origTime := time.Date(2020, 3, 21, 13, 00, 0, 0, time.UTC)
	tt = origTime
	var rate int64 = 100
	for i := int64(1); i < rate; i++ {
		tt = tt.Add(time.Millisecond * 100)
		if err := rl.Hit(context.TODO(), "user1", rate); err != nil {
			t.Fatal(err)
		}
	}
	tt = origTime.Add(time.Minute).Add(-1)
	if err := rl.Hit(context.TODO(), "user2", rate); err != nil {
		t.Fatal(err)
	}
	if err := rl.Hit(context.TODO(), "user1", rate); err == nil {
		t.Fatal("expected err")
	} else if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("got wrong err code: %v", status.Code(err))
	}
	tt = origTime.Add(time.Minute)
	if err := rl.Hit(context.TODO(), "user1", rate); err != nil {
		t.Fatal(err)
	}
}
