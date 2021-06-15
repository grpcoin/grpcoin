package realtimequote

import (
	"context"
	"testing"
	"time"
)

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
