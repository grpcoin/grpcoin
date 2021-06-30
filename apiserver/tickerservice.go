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

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/fanout"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tickerService struct {
	supportedTickers []string
	maxRate          time.Duration
	fanout           *fanout.QuoteFanoutService

	grpcoin.UnimplementedTickerInfoServer
}

func filterByProduct(ch <-chan realtimequote.Quote, product string) <-chan realtimequote.Quote {
	outCh := make(chan realtimequote.Quote)
	go func() {
		for m := range ch {
			if m.Product == product {
				outCh <- m
			}
		}
		close(outCh)
	}()
	return outCh
}

func (ts *tickerService) Watch(req *grpcoin.TickerWatchRequest, stream grpcoin.TickerInfo_WatchServer) error {
	// -USD suffix is now obsolete, keepin for back-compat with old clients

	req.Currency.Symbol = strings.TrimSuffix(req.GetCurrency().GetSymbol(), "-USD")

	if !realtimequote.IsSupported(ts.supportedTickers, req.Currency.Symbol) {
		return status.Errorf(codes.InvalidArgument, "only supported tickers are %#v", ts.supportedTickers)
	}
	ch, err := ts.fanout.RegisterWatch(stream.Context())
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to register ticker watch: %v", err))
	}
	ch = filterByProduct(ch, req.Currency.Symbol)
	ch = realtimequote.RateLimited(ch, ts.maxRate)
	for m := range ch {
		err = stream.Send(&grpcoin.Quote{
			T:     timestamppb.New(m.Time),
			Price: m.Price,
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			return err
		}
	}
	select {
	case <-stream.Context().Done():
		err := stream.Context().Err()
		if err == context.DeadlineExceeded || err == context.Canceled {
			return status.Error(codes.Canceled, fmt.Sprintf("client cancelled request: %v", err))
		}
		return status.Error(codes.Internal, fmt.Sprintf("unknown error on ctx: %v", err))
	default:
		return status.Error(codes.Internal, "failed to get prices, please retry by reconnecting")
	}
}
