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
	"log"
	"sync"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/gdax"
)

type QuoteProvider interface {
	// GetQuote provides real-time quote for ticker.
	// Can quit early if ctx is cancelled.
	GetQuote(ctx context.Context, ticker string) (*grpcoin.Amount, error)
}

type coinbaseQuoteProvider struct {
	ticker string

	lock        sync.RWMutex
	lastQuote   *grpcoin.Amount
	lastUpdated time.Time
}

const staleQuotePeriod = time.Second

func (cb *coinbaseQuoteProvider) GetQuote(ctx context.Context, _ string) (*grpcoin.Amount, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			cb.lock.RLock()
			if time.Since(cb.lastUpdated) > staleQuotePeriod {
				cb.lock.RUnlock()
				break
			}
			q := cb.lastQuote
			cb.lock.RUnlock()
			return q, nil
		}
		time.Sleep(time.Millisecond * 10) // TODO not so great but prevents 100% cpu
	}
}

// sync continually maintains a channel to Coinbase realtime prices.
// meant to be invoked in a goroutine
func (cb *coinbaseQuoteProvider) sync(ctx context.Context, ticker string) {
	for {
		if ctx.Err() != nil {
			return
		}
		ch, err := gdax.StartWatch(ctx, ticker+"-USD")
		if err != nil {
			// TODO plumb logger here
			log.Printf("warning: failed to connected to coinbase: %v", err)
			// TODO consider adding sleep/backoff to prevent DoSing coinbase API
			continue
		}
		for m := range ch {
			cb.lock.Lock()
			cb.lastUpdated = time.Now()
			cb.lastQuote = m.Price
			cb.lock.Unlock()
		}
		log.Printf("warn: coinbase channel terminated, reopening")
	}
}
