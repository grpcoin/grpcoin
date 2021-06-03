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
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/gdax"
)

type QuoteProvider struct {
	lock         sync.RWMutex
	lastBTCQuote *grpcoin.Amount
	lastETHQuote *grpcoin.Amount
	lastUpdated  time.Time

	Logger *zap.Logger
}

const staleQuotePeriod = time.Second * 2

func (cb *QuoteProvider) GetQuote(ctx context.Context, product string) (*grpcoin.Amount, error) {
	for {
		select {
		case <-ctx.Done():
			cb.lock.RLock()
			cb.Logger.Warn("quote request cancelled", zap.Error(ctx.Err()),
				zap.Duration("last_quote", time.Since(cb.lastUpdated)))
			cb.lock.RUnlock()
			return nil, ctx.Err()
		default:
			cb.lock.RLock()
			if time.Since(cb.lastUpdated) > staleQuotePeriod {
				cb.lock.RUnlock()
				break
			}
			btc,eth:=cb.lastBTCQuote,cb.lastETHQuote
			cb.lock.RUnlock()
			if product == "BTC" {
				return btc, nil
			} else if product == "ETH" {
				return eth, nil
			} else {
				return nil, fmt.Errorf("unknown trading product %q", product)
			}
		}
		time.Sleep(time.Millisecond * 10) // TODO not so great but prevents 100% cpu
	}
}

// sync continually maintains a channel to Coinbase realtime prices.
// meant to be invoked in a goroutine
func (cb *QuoteProvider) Sync(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		ch, err := gdax.StartWatch(ctx, "BTC-USD", "ETH-USD")
		if err != nil {
			cb.Logger.Warn("warning: failed to connected to coinbase", zap.Error(err))
			// TODO consider adding sleep/backoff to prevent DoSing coinbase API
			continue
		}
		for m := range ch {
			cb.lock.Lock()
			cb.lastUpdated = time.Now()
			if m.Product == "ETH-USD" {
				cb.lastETHQuote = m.Price
			} else if m.Product == "BTC-USD" {
				cb.lastBTCQuote = m.Price
			} else {
				cb.Logger.Warn("unknown product", zap.String("product", m.Product))
			}
			cb.lock.Unlock()
		}
		cb.Logger.Warn("coinbase channel terminated, reopening")
	}
}
