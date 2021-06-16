package common

import (
	"reflect"
	"testing"

	"github.com/grpcoin/grpcoin/api/grpcoin"
)

func TestParsePrice(t *testing.T) {
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
		{
			in:   "57469.123456789",
			want: &grpcoin.Amount{Units: 57_469, Nanos: 123_456_789},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParsePrice(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePrice(%s) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func BenchmarkParsePrice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParsePrice("123456.1234567")
	}
}
