package main

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ahmetb/grpcoin/api/grpcoin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWatch(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	grpcoin.RegisterTickerInfoServer(srv, new(tickerService))
	go srv.Serve(l)
	defer srv.Stop()
	defer l.Close()

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := grpcoin.NewTickerInfoClient(cc)
	count := 0
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*5)
	defer cleanup()

	stream, err := client.Watch(ctx, &grpcoin.Ticker{Ticker: "BTC-USD"})
	if err != nil {
		t.Fatal(err)
	}
	for {
		var m grpcoin.TickerQuote
		err = stream.RecvMsg(&m)
		if err != nil {
			if e := status.Convert(err); e != nil {
				if e.Code() == codes.DeadlineExceeded {
					break
				}
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
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	grpcoin.RegisterTickerInfoServer(srv, new(tickerService))
	go srv.Serve(l)
	defer srv.Stop()
	defer l.Close()

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	client := grpcoin.NewTickerInfoClient(cc)
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*5)
	defer cleanup()

	stream, err := client.Watch(ctx, &grpcoin.Ticker{Ticker: "BTC-USD"})
	if err != nil {
		t.Fatal(err)
	}
	for {
		var m grpcoin.TickerQuote
		err = stream.RecvMsg(&m)
		if err != nil {
			if e := status.Convert(err); e != nil {
				if e.Code() == codes.DeadlineExceeded {
					break
				}
			}
			panic(err)
		}
	}

	ctx, cleanup2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cleanup2()

	count := 0
	stream, err = client.Watch(ctx, &grpcoin.Ticker{Ticker: "BTC-USD"})
	if err != nil {
		t.Fatal(err)
	}
	for {
		var m grpcoin.TickerQuote
		err = stream.RecvMsg(&m)
		if err != nil {
			if e := status.Convert(err); e != nil {
				if e.Code() == codes.DeadlineExceeded {
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
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	grpcoin.RegisterTickerInfoServer(srv, new(tickerService))
	go srv.Serve(l)
	defer srv.Stop()
	defer l.Close()
	var wg sync.WaitGroup
	k := 20
	c := make([]int, k)
	for i := 0; i < k; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
			if err != nil {
				t.Fatal(err)
			}
			client := grpcoin.NewTickerInfoClient(cc)
			ctx, cleanup := context.WithTimeout(context.Background(), time.Second*5)
			defer cleanup()
			stream, err := client.Watch(ctx, &grpcoin.Ticker{Ticker: "BTC-USD"})
			if err != nil {
				t.Fatalf("routine %d: dial: %v", j, err)
			}
			for {
				var m grpcoin.TickerQuote
				err = stream.RecvMsg(&m)
				if err != nil {
					if e := status.Convert(err); e != nil {
						if e.Code() == codes.DeadlineExceeded {
							break
						}
					}
					t.Fatalf("routine %d: recv: %v", j, err)
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
