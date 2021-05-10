package gdax

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ahmetb/grpcoin/api/grpcoin"
)

func TestStartWatch(t *testing.T) {
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*3)
	defer cleanup()

	c, err := StartWatch(ctx, "BTC-USD")
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for {
		_, ok := <-c
		if !ok {
			break
		}
		count++
	}
	if count == 0 {
		t.Fatal("no messages received while the watch was on")
	}
	t.Logf("%d msgs received", count)
}

func TestRateLimited(t *testing.T) {
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*3)
	defer cleanup()

	in := make(chan Quote)
	go func() {
		for {
			if ctx.Err() != nil {
				close(in)
				return
			}
			in <- Quote{}
		}
	}()

	// 675*4=2700 < 3000
	out := RateLimited(in, time.Millisecond*675)
	count := 0
	for range out {
		count++
	}
	if expected := 5; count != expected {
		t.Fatalf("wrong msg recv count:%d expected:%d", count, expected)
	}
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
