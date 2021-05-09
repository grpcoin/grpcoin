package gdax

import (
	"context"
	"fmt"
	"log"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro/v2"
	"golang.org/x/time/rate"
)

type Quote struct {
	Product string    `json:"product"`
	Price   string    `json:"price"`
	Time    time.Time `json:"time"`
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
				Price: message.Price,
				Time:  message.Time.Time(),
			}
		}
	}()
	return ch, nil
}
