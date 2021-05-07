package main

import (
	"context"
	"log"
	"time"

	"github.com/ahmetb/paperbtctrade/gdax"
)

func main() {
	ch, err := gdax.StartWatch(context.Background(), "BTC-USD")
	if err != nil {
		panic(err)
	}
	ch = gdax.RateLimited(ch, time.Millisecond*5)
	log.SetFlags(log.Lmicroseconds)
	for v := range ch {
		log.Printf("[%s] %s: %s", v.Time, v.Product, v.Price)
	}
}
