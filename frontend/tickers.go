package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/grpcoin/grpcoin/gdax"
)

type tickerService struct {}

var newTicker tickerService

func wsTickers(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	ctx, stop := context.WithTimeout(context.Background(), time.Hour*5)

	quotes, err := gdax.StartWatch(ctx, "BTC-USD", "ETH-USD")
	if err != nil {
		stop()
		return
	}

	for m := range quotes {
		if err := conn.WriteJSON(&m); err != nil {
			conn.Close()
			if err == context.Canceled {
				break
			}
			return
		}
	}
	go client.reader()
	return
}
