package realtimequote

import (
	"context"

	"github.com/grpcoin/grpcoin/api/grpcoin"
)

type QuoteProvider interface {
	// GetQuote provides real-time quote for ticker.
	// Can quit early if ctx is cancelled.
	GetQuote(ctx context.Context, ticker string) (*grpcoin.Amount, error)
}
