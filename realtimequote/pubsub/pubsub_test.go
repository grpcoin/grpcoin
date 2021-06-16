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

package pubsub

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grpcoin/grpcoin/realtimequote"
)

func TestPubSub(t *testing.T) {
	src := make(chan realtimequote.Quote)

	var stopCalled int64
	stop := func() { atomic.AddInt64(&stopCalled, 1) }

	send := func(n int) {
		for i := 0; i < n; i++ {
			src <- realtimequote.Quote{}
			time.Sleep(time.Millisecond * 1) // if we send too fast, all is discarded
		}
	}

	bus := NewPubSub(src, stop)

	send(1)

	var recv1, recv2 int64
	ch1, ch2 := make(chan realtimequote.Quote), make(chan realtimequote.Quote)
	go func() {
		for range ch1 {
			atomic.AddInt64(&recv1, 1)
			t.Logf("recv1")
		}
	}()
	go func() {
		for range ch2 {
			atomic.AddInt64(&recv2, 1)
			t.Logf("recv2")
		}
	}()

	bus.Sub(ch1)
	send(5)

	bus.Sub(ch2)
	send(3)

	bus.Unsub(ch1)
	if _, ok := <-ch1; ok {
		t.Fatalf("ch1 should be closed")
	}
	if stopCalled != 0 {
		t.Fatalf("stop called %d times, expected no calls", stopCalled)
	}

	bus.Unsub(ch2)
	if _, ok := <-ch2; ok {
		t.Fatalf("ch2 should be closed")
	}
	stopCalls := atomic.LoadInt64(&stopCalled)
	if stopCalls != 1 {
		t.Fatalf("stop called %d times, expected 1", stopCalls)
	}

	if atomic.LoadInt64(&recv1) == 0 {
		t.Fatalf("received nothing at ch1")
	}
	if atomic.LoadInt64(&recv2) == 0 {
		t.Fatalf("received nothing at ch2")
	}
}

func TestPubSubOnSourceClose(t *testing.T) {
	src := make(chan realtimequote.Quote)

	var mu sync.Mutex
	var stopCalled bool
	stop := func() {
		mu.Lock()
		stopCalled = true
		mu.Unlock()
	}
	bus := NewPubSub(src, stop)
	ch1, ch2 := make(chan realtimequote.Quote), make(chan realtimequote.Quote)
	bus.Sub(ch1)
	bus.Sub(ch2)

	close(src) // i.e. data stream disconnected
	if _, ok := <-ch1; ok {
		t.Fatalf("ch1 should be closed")
	}
	if _, ok := <-ch2; ok {
		t.Fatalf("ch2 should be closed")
	}
	mu.Lock()
	defer mu.Unlock()
	if !stopCalled {
		t.Fatalf("stop() should have been called")
	}
}
