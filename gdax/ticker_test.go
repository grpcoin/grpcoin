package gdax

import (
	"context"
	"testing"
	"time"
)

func TestStartWatch(t *testing.T) {
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*3)
	defer cleanup()

	c, err := StartWatch(ctx, "BTC-USD")
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for {
		v, ok := <-c
		if !ok {
			break
		}
		if v.Price == "" {
			t.Fatalf("empty price received on count %d", count)
		}
		count++
	}
	if count == 0 {
		t.Fatal("no messages received while the watch was on")
	}
	t.Logf("%d msgs received", count)
}

func TestRateLimited(t *testing.T) {
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*3)
	defer cleanup()

	in := make(chan Quote)
	go func() {
		for {
			if ctx.Err() != nil {
				close(in)
				return
			}
			in <- Quote{}
		}
	}()

	// 675*4=2700 < 3000
	out := RateLimited(in, time.Millisecond*675)
	count := 0
	for range out {
		count++
	}
	if expected := 5; count != expected {
		t.Fatalf("wrong msg recv count:%d expected:%d", count, expected)
	}
}
