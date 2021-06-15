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

package realtimequote

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	src := make(chan Quote)

	var stopCalled int64
	stop := func() { atomic.AddInt64(&stopCalled, 1) }

	send := func(n int) {
		for i := 0; i < n; i++ {
			src <- Quote{}
			time.Sleep(time.Millisecond * 1) // if we send too fast, all is discarded
			t.Logf("sent%d", i)
		}
	}

	bus := NewPubSub(src, stop)

	send(1)

	var recv1, recv2 int
	ch1, ch2 := make(chan Quote), make(chan Quote)
	go func() {
		for range ch1 {
			recv1++
			t.Logf("recv1")
		}
	}()
	go func() {
		for range ch2 {
			recv2++
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
	if stopCalled != 1 {
		t.Fatalf("stop called %d times, expected 1", stopCalled)
	}

	if recv1 == 0 {
		t.Fatalf("received nothing at ch1")
	}
	if recv2 == 0 {
		t.Fatalf("received nothing at ch2")
	}
}

func TestPubSubOnSourceClose(t *testing.T) {
	src := make(chan Quote)

	var stopCalled bool
	stop := func() { stopCalled = true }
	bus := NewPubSub(src, stop)
	ch1, ch2 := make(chan Quote), make(chan Quote)
	bus.Sub(ch1)
	bus.Sub(ch2)

	close(src) // i.e. data stream disconnected
	if _, ok := <-ch1; ok {
		t.Fatalf("ch1 should be closed")
	}
	if _, ok := <-ch2; ok {
		t.Fatalf("ch2 should be closed")
	}
	if !stopCalled {
		t.Fatalf("stop() should have been called")
	}
}
