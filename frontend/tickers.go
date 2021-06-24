package main

import (
	"context"
	"net/http"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/realtimequote"
	"go.uber.org/zap"
)

func (fe *frontend) wsTickers(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ch, err := fe.QuoteFanout.RegisterWatch(ctx)
	if err != nil {
		return err
	}

	ch = realtimequote.PerSymbolRateLimited(ch, time.Second)

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
