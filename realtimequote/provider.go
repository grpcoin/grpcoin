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
	"context"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
)

type Quote struct {
	Product string
	Price   *grpcoin.Amount
	Time    time.Time
}

type QuoteProvider interface {
	// GetQuote provides real-time quote for ticker (e.g. BTC, ETH, DOGE, ...).
	// It can block until it gets a "recent enough" quote. Can quit early if ctx is cancelled.
	GetQuote(ctx context.Context, ticker string) (*grpcoin.Amount, error)
}

type QuoteStream interface {
	// Watch provides real-time quotes for given product (e.g. BTC, ETH, DOGE, ...)
	// err is returned if it fails to connect and start streaming.
	// ch is closed when ctx is done, or if the stream disconnects.
	Watch(ctx context.Context, products ...string) (ch <-chan Quote, err error)
}

type QuoteStreamFunc func(ctx context.Context, products ...string) (ch <-chan Quote, err error)

func (f QuoteStreamFunc) Watch(ctx context.Context, products ...string) (ch <-chan Quote, err error) {
	return f(ctx, products...)
}
