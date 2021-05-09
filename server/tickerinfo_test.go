package main

import (
	"context"
	"net"
	"reflect"
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

	ctx, cleanup = context.WithTimeout(context.Background(), time.Second*5)
	defer cleanup()

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

func Test_convertPrice(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *grpcoin.Amount
	}{
		{
			in:   "",
			want: &grpcoin.Amount{},
		},
		{
			in:   "0.0",
			want: &grpcoin.Amount{},
		},
		{
			in:   "3.",
			want: &grpcoin.Amount{Units: 3},
		},
		{
			in:   ".3",
			want: &grpcoin.Amount{Nanos: 300_000_000},
		},
		{
			in:   "0.072",
			want: &grpcoin.Amount{Nanos: 72_000_000},
		},
		{
			in:   "57469.71",
			want: &grpcoin.Amount{Units: 57_469, Nanos: 710_000_000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertPrice(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertPrice(%s) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
