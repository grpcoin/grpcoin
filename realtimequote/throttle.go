package realtimequote

import (
	"time"

	"golang.org/x/time/rate"
)

// RateLimited blocks output from the channel to send no more than one quote
// per d.
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
