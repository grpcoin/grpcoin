package main

import (
	"sync"

	"github.com/ahmetb/grpcoin/gdax"
)

type PubSub struct {
	// Src is where the producer sends the messages. Close
	// this to close registered subscribers.
	src <-chan gdax.Quote

	// stop is called when the last client unregisters.
	stop func()

	mu   sync.Mutex
	subs map[chan<- gdax.Quote]bool
}

// NewPubSub returns an in-memory pubsub topic.
// Messages are read from src. If src closes, subscribers will be closed.
// When the last subscriber is unsubscribed, stop is called.
func NewPubSub(src <-chan gdax.Quote, stop func()) *PubSub {
	p := &PubSub{src: src, stop: stop,
		subs: make(map[chan<- gdax.Quote]bool)}
	go p.fanout()
	return p
}

// Sub creates a subscription that pushes to ch.
// If src closes, ch will be closed.
// If message blocks from being sent on ch, it will be dropped.
func (p *PubSub) Sub(ch chan<- gdax.Quote) {
	p.mu.Lock()
	p.subs[ch] = true
	p.mu.Unlock()
}

// Unsub removes subscription and closes ch.
func (p *PubSub) Unsub(ch chan<- gdax.Quote) {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.subs[ch]
	if !ok {
		return
	}
	delete(p.subs, ch)
	close(ch)
	if len(p.subs) == 0 {
		p.stop()
	}
}

func (p *PubSub) fanout() {
	for m := range p.src {
		p.mu.Lock()
		for c := range p.subs {
			select {
			case c <- m:
			default: // drop message
			}
		}
		p.mu.Unlock()
	}
	// if pub ch closes, we close subscribers
	p.mu.Lock()
	for c := range p.subs {
		delete(p.subs, c)
		close(c)
	}
	p.stop()
	p.mu.Unlock()
}
