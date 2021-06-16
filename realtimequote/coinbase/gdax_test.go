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

package coinbase

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
)

func TestStartWatch(t *testing.T) {
	if testing.Short() {
		t.Skip("makes calls to coinbase")
	}
	ctx, cleanup := context.WithTimeout(context.Background(), time.Second*3)
	defer cleanup()

	c, err := StartWatch(ctx, "BTC")
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for {
		v, ok := <-c
		if !ok {
			break
		}
		if strings.HasSuffix(v.Product, "-USD") {
			t.Fatalf("quote has -USD suffix %#v", v)
		}
		count++
	}
	if count == 0 {
		t.Fatal("no messages received while the watch was on")
	}
	t.Logf("%d msgs received", count)
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