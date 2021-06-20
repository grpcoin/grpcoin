package main

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (fe *frontend) wsTickers(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.Error(err)
	}
	defer conn.Close()

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	for {
		<-time.After(2 * time.Second)
		quotes, err := fe.getQuotes(context.Background())
		if err != nil {
			zap.Error(err)
		}

		if err := conn.WriteJSON(&quotes); err != nil {
			conn.Close()
		}
	}
}
