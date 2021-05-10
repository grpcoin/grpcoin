package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ahmetb/grpcoin/api/grpcoin"
	"github.com/ahmetb/grpcoin/gdax"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tickerService struct {
	grpcoin.UnimplementedTickerInfoServer

	lock sync.Mutex
	bus  *PubSub
}

func (t *tickerService) initWatch() error {
	fmt.Println("init")
	t.lock.Lock()
	if t.bus != nil {
		t.lock.Unlock()
		return nil
	}

	fmt.Println("connecting to ws")
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		t.lock.Lock()
		t.bus = nil
		fmt.Println("last client disconnected")
		t.lock.Unlock()
	}()
	quotes, err := gdax.StartWatch(ctx, "BTC-USD")
	if err != nil {
		t.lock.Unlock()
		return err
	}
	quotes = gdax.RateLimited(quotes, time.Millisecond*500)
	t.bus = NewPubSub(quotes, stop)
	t.lock.Unlock()
	return nil
}

func (t *tickerService) registerWatch(ctx context.Context) (<-chan gdax.Quote, error) {
	ch := make(chan gdax.Quote)
	if err := t.initWatch(); err != nil {
		return nil, err
	}
	t.bus.Sub(ch)
	go func() {
		<-ctx.Done()
		fmt.Println("rpc watch done")
		t.bus.Unsub(ch)
	}()
	return ch, nil
}

func (f *tickerService) Watch(req *grpcoin.Ticker, stream grpcoin.TickerInfo_WatchServer) error {
	if req.GetTicker() != "BTC-USD" {
		return status.Error(codes.InvalidArgument, "only supported ticker is BTC-USD")
	}
	ch, err := f.registerWatch(stream.Context())
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to register ticker watch: %v", err))
	}
	for m := range ch {
		err = stream.Send(&grpcoin.TickerQuote{
			T:     timestamppb.New(m.Time),
			Price: convertPrice(m.Price),
		})
		if err != nil {
			return err
		}
	}
	return status.Error(codes.Internal, "failed to get prices, please retry by reconnecting")
}

func convertPrice(p string) *grpcoin.Amount {
	// TODO this is currently inefficient as each connected client converts the
	// amount on every msg. move this to gdax package.
	out := strings.SplitN(p, ".", 2)
	if len(out) == 0 {
		return &grpcoin.Amount{}
	}
	if out[0] == "" {
		out[0] = "0"
	}
	i, _ := strconv.ParseInt(out[0], 10, 64)
	if len(out) == 1 {
		return &grpcoin.Amount{Units: i}
	}
	out[1] += strings.Repeat("0", 9-len(out[1]))
	j, _ := strconv.Atoi(out[1])
	return &grpcoin.Amount{Units: i, Nanos: int32(j)}
}
