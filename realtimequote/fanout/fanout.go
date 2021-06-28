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

package fanout

import (
	"context"
	"sync"

	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/pubsub"
)

type QuoteFanoutService struct {
	lock                   sync.Mutex
	bus                    *pubsub.PubSub
	quoteStreamInitializer func(ctx context.Context) (<-chan realtimequote.Quote, error)
}

func NewQuoteFanoutService(q func(ctx context.Context) (<-chan realtimequote.Quote, error)) *QuoteFanoutService {
	return &QuoteFanoutService{quoteStreamInitializer: q}
}

func (q *QuoteFanoutService) initWatch() error {
	q.lock.Lock()
	if q.bus != nil {
		q.lock.Unlock()
		return nil
	}

	ctx, stop := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		q.lock.Lock()
		q.bus = nil
		q.lock.Unlock()
	}()
	quotes, err := q.quoteStreamInitializer(ctx)
	if err != nil {
		stop()
		q.lock.Unlock()
		return err
	}
	q.bus = pubsub.NewPubSub(quotes, stop)
	q.lock.Unlock()
	return nil
}

func (q *QuoteFanoutService) RegisterWatch(ctx context.Context) (<-chan realtimequote.Quote, error) {
	ch := make(chan realtimequote.Quote)
	if err := q.initWatch(); err != nil {
		return nil, err
	}
	q.bus.Sub(ch)
	go func() {
		<-ctx.Done()
		q.bus.Unsub(ch)
	}()
	return ch, nil
}
