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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/fanout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func prepTickerService(t *testing.T) *grpc.ClientConn {
	t.Helper()
	l := bufconn.Listen(1024)
	srv := grpc.NewServer()
	ts := &tickerService{
		maxRate:          time.Millisecond,
		supportedTickers: []string{"BTC"},
		fanout: fanout.NewQuoteFanoutService(func(ctx context.Context) (<-chan realtimequote.Quote, error) {
			return (mockQuoteStream{
				product: "BTC",
				n:       10,
				price:   &grpcoin.Amount{Units: 35_000}}).Watch(ctx)
		}),
	}
	grpcoin.RegisterTickerInfoServer(srv, ts)
	go srv.Serve(l)
	t.Cleanup(func() {
		srv.Stop()
		l.Close()
	})
	cc, err := grpc.DialContext(context.TODO(), "bufnet", grpc.WithInsecure(), grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
		return l.Dial()
	}))
	if err != nil {
		t.Fatal(err)
	}
	return cc
}

func TestWatch(t *testing.T) {
	client := grpcoin.NewTickerInfoClient(prepTickerService(t))
	count := 0
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second)
	defer cleanup()

	stream, err := client.Watch(ctx,
		&grpcoin.TickerWatchRequest{Currency: &grpcoin.Currency{Symbol: "BTC"}})
	if err != nil {
		t.Fatal(err)
	}
	for {
		var m grpcoin.Quote
		err = stream.RecvMsg(&m)
		if err != nil {
			if status.Code(err) == codes.DeadlineExceeded || status.Code(err) == codes.Canceled {
				break
			}
			t.Logf("received so far %d", count)
			t.Fatal(err)
		}
		count++
	}
	t.Logf("received so far %d", count)
	if count == 0 {
		t.Fatal("no msgs received")
	}
}

func TestWatchReconnect(t *testing.T) {
	client := grpcoin.NewTickerInfoClient(prepTickerService(t))
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second)
	defer cleanup()

	stream, err := client.Watch(ctx,
		&grpcoin.TickerWatchRequest{Currency: &grpcoin.Currency{Symbol: "BTC"}})
	if err != nil {
		t.Fatal(err)
	}
	for {
		var m grpcoin.Quote
		err = stream.RecvMsg(&m)
		if err != nil {
			if e := status.Convert(err); e != nil {
				if status.Code(err) == codes.DeadlineExceeded || status.Code(err) == codes.Canceled {
					break
				}
			}
			panic(err)
		}
	}

	ctx, cleanup2 := context.WithTimeout(context.Background(), time.Second)
	defer cleanup2()

	count := 0
	stream, err = client.Watch(ctx,
		&grpcoin.TickerWatchRequest{Currency: &grpcoin.Currency{Symbol: "BTC"}})
	if err != nil {
		t.Fatal(err)
	}
	for {
		var m grpcoin.Quote
		err = stream.RecvMsg(&m)
		if err != nil {
			if e := status.Convert(err); e != nil {
				if e.Code() == codes.DeadlineExceeded || e.Code() == codes.Canceled {
					break
				}
			}
			t.Fatal(err)
		}
		count++
	}
	t.Logf("received so far %d", count)
	if count == 0 {
		t.Fatal("no msgs received")
	}
}

func TestWatchMulti(t *testing.T) {
	client := grpcoin.NewTickerInfoClient(prepTickerService(t))
	var wg sync.WaitGroup
	k := 20
	c := make([]int, k)
	for i := 0; i < k; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			ctx, cleanup := context.WithTimeout(context.Background(), time.Second)
			defer cleanup()
			stream, err := client.Watch(ctx,
				&grpcoin.TickerWatchRequest{Currency: &grpcoin.Currency{Symbol: "BTC"}})
			if err != nil {
				panic(fmt.Sprintf("routine %d: dial: %v", j, err))
			}
			for {
				var m grpcoin.Quote
				err = stream.RecvMsg(&m)
				if err != nil {
					if status.Code(err) == codes.DeadlineExceeded || status.Code(err) == codes.Canceled {
						break
					}
					panic(fmt.Sprintf("routine %d: recv: %v", j, err))
				}
				c[j]++
			}
		}(i)
	}
	wg.Wait()
	for i := 0; i < k; i++ {
		if c[i] == 0 {
			t.Fatalf("count=0 on routine%d", i)
		}
	}
	t.Logf("%#v", c)
}
