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
	"fmt"
	"log"
	"strings"
	"time"

	gobinance "github.com/adshao/go-binance/v2"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/common"
)

func WatchSymbols(ctx context.Context, products ...string) (<-chan realtimequote.Quote, error) {
	gobinance.WebsocketKeepalive = true // handle sending pong frames
	symbols := make([]string, len(products))
	for i, s := range products {
		symbols[i] = strings.ToLower(s + "USDT")
	}

	out := make(chan realtimequote.Quote)
	doneC, stopC, err := gobinance.WsCombinedAggTradeServe(symbols, func(event *gobinance.WsAggTradeEvent) {
		out <- realtimequote.Quote{
			Product: strings.ToUpper(strings.TrimSuffix(event.Symbol, "USDT")),
			Price:   common.ParsePrice(event.Price),
			Time:    time.Unix(event.Time/1000, event.Time%1000*1_000_000)}
	}, func(err error) {
		log.Print(err)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect binance: %w", err)
	}
	go func() {
		<-ctx.Done()
		stopC <- struct{}{}
	}()
	go func() {
		<-doneC
		close(out)
	}()
	return out, nil
}
