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

package binance

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/grpcoin/grpcoin/realtimequote"
)

func TestWatch(t *testing.T) {
	qf := realtimequote.QuoteStreamFunc(WatchSymbols)
	timedCtx, done := context.WithTimeout(context.Background(), time.Second*6)
	defer done()
	ctx, cancel := context.WithCancel(timedCtx)
	defer cancel()

	ch, err := qf.Watch(ctx, "BTC", "DOGE", "ETH", "DOT")
	if err != nil {
		t.Fatal(err)
	}
	m := map[string][]realtimequote.Quote{
		"DOGE": nil,
		"BTC":  nil,
		"ETH":  nil,
		"DOT":  nil,
	}

	recvQuotesForAll := false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for q := range ch {
			if q.Price.Units == 0 && q.Price.Nanos == 0 {
				t.Fatalf("zero price recvd on Quote:%#v", q)
			}
			if x := time.Now().Sub(q.Time); x > 5*time.Second || x < -5*time.Second {
				t.Fatalf("t=%v (%s) > +-5 seconds", x, q.Time)
			}
			m[q.Product] = append(m[q.Product], q)
			allReady := true
			for _, v := range m {
				allReady = allReady && len(v) > 0
			}
			if allReady {
				recvQuotesForAll = true
				break
			}
		}
	}()
	wg.Wait()
	if !recvQuotesForAll {
		for k, v := range m {
			t.Logf("%s %d quotes", k, len(v))
		}
		t.Fatal("did not recv quotes for all")
	}
}
