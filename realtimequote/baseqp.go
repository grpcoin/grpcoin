package realtimequote

import (
	"context"
	"sync"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"go.uber.org/zap"
)

type quote struct {
	amount      *grpcoin.Amount
	lastUpdated time.Time
}

type ReconnectingQuoteProvider struct {
	logger           *zap.Logger
	products         []string
	qs               QuoteStream
	staleQuotePeriod time.Duration

	lock   sync.RWMutex
	quotes map[string]quote
}

const DefaultStaleQuotePeriod = time.Second * 10 // fails if quote is older than this
const DefaultReconnectInterval = time.Millisecond * 100

// NewReconnectingQuoteProvider maintains a auto-reconnecting stream to the
// quote stream provider, and closes the underlying stream when ctx is done
// and stops re-connecting.
func NewReconnectingQuoteProvider(ctx context.Context, log *zap.Logger, quoteStream QuoteStream, products ...string) QuoteProvider {
	qp := &ReconnectingQuoteProvider{
		logger:           log,
		staleQuotePeriod: DefaultStaleQuotePeriod,
		qs:               quoteStream,
		products:         products,
	}
	go qp.sync(ctx)
	return qp
}

func (qp *ReconnectingQuoteProvider) GetQuote(ctx context.Context, product string) (*grpcoin.Amount, error) {
	stalePeriod := qp.staleQuotePeriod
	if stalePeriod == 0 {
		stalePeriod = DefaultStaleQuotePeriod
	}
	for {
		select {
		case <-ctx.Done():
			qp.logger.Warn("quote request cancelled", zap.Error(ctx.Err()))
			return nil, ctx.Err()
		default:
			qp.lock.RLock()
			q := qp.quotes[product]
			if time.Since(q.lastUpdated) > stalePeriod {
				qp.lock.RUnlock()
				break
			}
			amount := q.amount
			qp.lock.RUnlock()
			return amount, nil
		}
		time.Sleep(time.Millisecond * 10) // TODO not so great but prevents 100% cpu
	}
}

// sync keeps track of incoming quotes from QuoteStream to provide them via
// GetQuote. It is meant to be invoked in a goroutine. Closes the underlying
// quote stream if ctx is done.
func (qp *ReconnectingQuoteProvider) sync(ctx context.Context) {
	qp.lock.Lock()
	qp.quotes = make(map[string]quote)
	qp.lock.Unlock()
	for {
		if ctx.Err() != nil {
			return
		}
		ch, err := qp.qs.Watch(ctx, qp.products...)
		if err != nil {
			qp.logger.Warn("warning: failed to connected to quote stream", zap.Error(err))
			time.Sleep(DefaultReconnectInterval)
			continue
		}
		for m := range ch {
			qp.lock.Lock()
			q := qp.quotes[m.Product]
			q.amount = m.Price
			q.lastUpdated = time.Now()
			qp.quotes[m.Product] = q
			qp.lock.Unlock()
		}
		qp.logger.Warn("quote stream broken, reopening")
	}
}
