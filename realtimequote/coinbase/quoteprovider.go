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

package coinbase

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/gdax"
	"github.com/grpcoin/grpcoin/realtimequote"
)

type quote struct {
	amount      *grpcoin.Amount
	lastUpdated time.Time
}

type QuoteProvider struct {
	Logger *zap.Logger

	lock   sync.RWMutex
	quotes map[string]quote
}

const staleQuotePeriod = time.Second * 2

func (cb *QuoteProvider) GetQuote(ctx context.Context, product string) (*grpcoin.Amount, error) {
	product += "-USD"
	for {
		select {
		case <-ctx.Done():
			cb.lock.RLock()
			cb.Logger.Warn("quote request cancelled", zap.Error(ctx.Err()))
			cb.lock.RUnlock()
			return nil, ctx.Err()
		default:
			cb.lock.RLock()
			q := cb.quotes[product]
			if time.Since(q.lastUpdated) > staleQuotePeriod {
				cb.lock.RUnlock()
				break
			}
			amount := q.amount
			cb.lock.RUnlock()
			return amount, nil
		}
		time.Sleep(time.Millisecond * 10) // TODO not so great but prevents 100% cpu
	}
}

// sync continually maintains a channel to Coinbase realtime prices.
// meant to be invoked in a goroutine
func (cb *QuoteProvider) Sync(ctx context.Context) {
	cb.lock.Lock()
	cb.quotes = make(map[string]quote)
	cb.lock.Unlock()
	for {
		if ctx.Err() != nil {
			return
		}
		ch, err := gdax.StartWatch(ctx, realtimequote.SupportedProducts...)
		if err != nil {
			cb.Logger.Warn("warning: failed to connected to coinbase", zap.Error(err))
			// TODO consider adding sleep/backoff to prevent DoSing coinbase API
			continue
		}
		for m := range ch {
			cb.lock.Lock()
			q := cb.quotes[m.Product]
			q.amount = m.Price
			q.lastUpdated = time.Now()
			cb.quotes[m.Product] = q
			cb.lock.Unlock()
		}
		cb.Logger.Warn("coinbase channel terminated, reopening")
	}
}
