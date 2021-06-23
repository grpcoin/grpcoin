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
