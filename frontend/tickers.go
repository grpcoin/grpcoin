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
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/realtimequote"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (fe *frontend) wsTickers(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ch, err := fe.QuoteFanout.RegisterWatch(ctx)
	if err != nil {
		return err
	}

	ch = realtimequote.RateLimited(ch, time.Second*3)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	type qr struct {
		Ticker string          `json:"t"`
		Price  *grpcoin.Amount `json:"p"`
	}

	successiveWriteErrs := 0
	for q := range ch {
		if successiveWriteErrs >= 3 {
			conn.Close()
			loggerFrom(r.Context()).Debug("disconnecting client", zap.Error(err))
			return nil
		}
		if err := conn.WriteJSON(qr{Ticker: q.Product, Price: q.Price}); err != nil {
			successiveWriteErrs++
			loggerFrom(r.Context()).Debug("ws write failed", zap.Error(err))
		} else {
			successiveWriteErrs = 0
		}
	}
	return nil
}
