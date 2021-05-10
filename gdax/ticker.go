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

package gdax

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ahmetb/grpcoin/api/grpcoin"
	ws "github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro/v2"
	"golang.org/x/time/rate"
)

type Quote struct {
	Product string          `json:"product"`
	Price   *grpcoin.Amount `json:"price"`
	Time    time.Time       `json:"time"`
}

func RateLimited(ch <-chan Quote, d time.Duration) <-chan Quote {
	out := make(chan Quote)
	lim := rate.Every(d)
	l := rate.NewLimiter(lim, 1)
	go func() {
		for m := range ch {
			if l.Allow() {
				out <- m
			}
			continue
		}
		close(out)
	}()
	return out
}

func StartWatch(ctx context.Context, product string) (<-chan Quote, error) {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	subscribe := coinbasepro.Message{
		Type: "subscribe",
		Channels: []coinbasepro.MessageChannel{
			{
				Name: "ticker",
				ProductIds: []string{
					product,
				},
			},
		},
	}

	if err := wsConn.WriteJSON(subscribe); err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	ch := make(chan Quote)
	go func() {
		for {
			if ctx.Err() != nil {
				close(ch)
				return
			}
			var message coinbasepro.Message
			if err := wsConn.ReadJSON(&message); err != nil {
				log.Printf("warn: json read/parse err: %v", err)
				close(ch)
				return
			}
			if message.Type != "ticker" {
				continue
			}
			ch <- Quote{Product: product,
				Price: convertPrice(message.Price),
				Time:  message.Time.Time(),
			}
		}
	}()
	return ch, nil
}

func convertPrice(p string) *grpcoin.Amount {
	out := strings.SplitN(p, ".", 2)
	if len(out) == 0 {
		return &grpcoin.Amount{}
	}
	if out[0] == "" {
		out[0] = "0"
	}
	i, _ := strconv.ParseInt(out[0], 10, 64)
	if len(out) == 1 {
		return &grpcoin.Amount{Units: i}
	}
	out[1] += strings.Repeat("0", 9-len(out[1]))
	j, _ := strconv.Atoi(out[1])
	return &grpcoin.Amount{Units: i, Nanos: int32(j)}
}
