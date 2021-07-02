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
	"testing"
	"time"

	"github.com/grpcoin/grpcoin/testutil"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRateLimiter_Hit(t *testing.T) {
	rc := testutil.MockRedis(t)

	var tt time.Time
	rl := New(rc,
		func() time.Time { return tt },
		trace.NewNoopTracerProvider().Tracer(""),
		time.Minute)

	origTime := time.Date(2020, 3, 21, 13, 00, 0, 0, time.UTC)
	tt = origTime
	var rate int64 = 100
	for i := int64(0); i < rate; i++ {
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
