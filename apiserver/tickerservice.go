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
	"fmt"
	"sync"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/coinbase/gdax"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tickerService struct {
	grpcoin.UnimplementedTickerInfoServer

	lock sync.Mutex
	bus  *realtimequote.PubSub
}

func (t *tickerService) initWatch() error {
	t.lock.Lock()
	if t.bus != nil {
		t.lock.Unlock()
		return nil
	}

	ctx, stop := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		t.lock.Lock()
		t.bus = nil
		t.lock.Unlock()
	}()
	quotes, err := gdax.StartWatch(ctx, realtimequote.SupportedProducts...)
	if err != nil {
		stop()
		t.lock.Unlock()
		return err
	}
	t.bus = realtimequote.NewPubSub(quotes, stop)
	t.lock.Unlock()
	return nil
}

func (t *tickerService) registerWatch(ctx context.Context) (<-chan realtimequote.Quote, error) {
	ch := make(chan realtimequote.Quote)
	if err := t.initWatch(); err != nil {
		return nil, err
	}
	t.bus.Sub(ch)
	go func() {
		<-ctx.Done()
		t.bus.Unsub(ch)
	}()
	return ch, nil
}

func filterProduct(ch <-chan realtimequote.Quote, product string) <-chan realtimequote.Quote {
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

func (f *tickerService) Watch(req *grpcoin.QuoteTicker, stream grpcoin.TickerInfo_WatchServer) error {
	if !realtimequote.IsSupported(realtimequote.SupportedProducts, req.GetTicker()) {
		return status.Errorf(codes.InvalidArgument, "only supported tickers are %#v", realtimequote.SupportedProducts)
	}
	ch, err := f.registerWatch(stream.Context())
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to register ticker watch: %v", err))
	}
	ch = filterProduct(ch, req.GetTicker())
	ch = realtimequote.RateLimited(ch, time.Millisecond*300)
	for m := range ch {
		err = stream.Send(&grpcoin.Quote{
			T:     timestamppb.New(m.Time),
			Price: m.Price,
		})
		if err != nil {
			if err == context.Canceled {
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
