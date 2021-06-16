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

package pubsub

import (
	"sync"

	"github.com/grpcoin/grpcoin/realtimequote"
)

type PubSub struct {
	// src is where the producer sends the messages. Close
	// this to close registered subscribers.
	src <-chan realtimequote.Quote

	// stop is called when the last client unregisters.
	stop func()

	mu   sync.Mutex
	subs map[chan<- realtimequote.Quote]bool
}

// NewPubSub returns an in-memory pubsub topic.
// Messages are read from src. If src closes, subscribers will be closed.
// When the last subscriber is unsubscribed, stop is called.
func NewPubSub(src <-chan realtimequote.Quote, stop func()) *PubSub {
	p := &PubSub{src: src, stop: stop,
		subs: make(map[chan<- realtimequote.Quote]bool)}
	go p.fanout()
	return p
}

// Sub creates a subscription that pushes to ch.
// If src closes, ch will be closed.
// If message blocks from being sent on ch, it will be dropped.
func (p *PubSub) Sub(ch chan<- realtimequote.Quote) {
	p.mu.Lock()
	p.subs[ch] = true
	p.mu.Unlock()
}

// Unsub removes subscription and closes ch.
func (p *PubSub) Unsub(ch chan<- realtimequote.Quote) {
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
