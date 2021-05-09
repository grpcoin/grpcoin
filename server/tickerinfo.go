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

	mu   sync.Mutex
	outs []reg

	mu2      sync.Mutex
	tickerCh <-chan gdax.Quote

	stop func()
}

type reg struct {
	ctx context.Context
	out chan<- gdax.Quote
}

func (t *tickerService) initWatch() error {
	t.mu2.Lock()
	if t.tickerCh != nil {
		t.mu2.Unlock()
		return nil
	}

	streamCtx, cancel := context.WithCancel(context.Background())
	fmt.Println("connecting to ws")
	c, err := gdax.StartWatch(streamCtx, "BTC-USD")
	if err != nil {
		t.mu2.Unlock()
		return err
	}

	t.tickerCh = gdax.RateLimited(c, time.Millisecond*500)
	t.stop = cancel
	t.mu2.Unlock()

	go func() {
		t.mu2.Lock()
		readCh := t.tickerCh
		t.mu2.Unlock()
		for m := range readCh {
			t.mu.Lock()
			outs := t.outs
			t.mu.Unlock()
			for _, o := range outs {
				select {
				case <-o.ctx.Done():
					// subscriber quit, unregister
					// TODO extract subscription mgmt to its own goroutine
					t.mu.Lock()
					newOuts := make([]reg, 0, len(t.outs)-1)
					for i := 0; i < len(t.outs); i++ {
						if t.outs[i] != o {
							newOuts = append(newOuts, t.outs[i])
						}
					}
					t.outs = newOuts
					if len(newOuts) == 0 {
						// no more clients streaming
						// TODO this is currently not working
						t.stop()
						fmt.Println("last client disconnected")
					}
					close(o.out)
					t.mu.Unlock()
				default:
					o.out <- m
				}
			}
		}

		// ws conn closed
		fmt.Println("ticker ch closed")
		t.mu.Lock()
		t.mu2.Lock()
		t.tickerCh = nil
		for _, o := range t.outs {
			close(o.out)
		}
		t.outs = nil
		t.mu2.Unlock()
		t.mu.Unlock()
	}()
	return nil
}

func (t *tickerService) registerWatch(ctx context.Context) (<-chan gdax.Quote, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.initWatch(); err != nil {
		return nil, err
	}
	ch := make(chan gdax.Quote)
	t.outs = append(t.outs, reg{ctx, ch})
	// TODO add goroutine here to unregister this on ctx.Done
	return ch, nil
}

func (f *tickerService) Watch(req *grpcoin.Ticker, stream grpcoin.TickerInfo_WatchServer) error {
	if req.GetTicker() != "BTC-USD" {
		return status.Error(codes.InvalidArgument, "only supported ticker is BTC-USD")
	}
	go func() {
		<-stream.Context().Done()
		fmt.Println("client disconnected (from rpc impl)")
	}()
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
