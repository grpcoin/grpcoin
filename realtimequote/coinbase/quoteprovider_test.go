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
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCoinbaseQuoteProvider_GetQuote(t *testing.T) {
	if testing.Short() {
		t.Skip("makes calls to coinbase")
	}
	lg, _ := zap.NewDevelopment()
	cb := &QuoteProvider{Logger: lg}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go cb.Sync(ctx)

	q1, err := cb.GetQuote(ctx, "BTC")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	ctx2, done := context.WithTimeout(ctx, time.Second)
	defer done()
	q2, err := cb.GetQuote(ctx2, "BTC")
	if err != nil {
		t.Fatal(err)
	}

	if q1.String() == q2.String() {
		t.Fatal("identical quotes")
	}

	_, err = cb.GetQuote(context.TODO(), "ETH")
	if err != nil {
		t.Fatal(err)
	}
}
