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

package userdb

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_toDecimal(t *testing.T) {
	tests := []struct {
		a    *grpcoin.Amount
		want string
	}{
		{
			a:    &grpcoin.Amount{Units: 3, Nanos: 3},
			want: ("3.000000003"),
		},
		{
			a:    &grpcoin.Amount{Units: -3, Nanos: -3},
			want: ("-3.000000003"),
		},
		{
			a:    &grpcoin.Amount{Units: 3, Nanos: 300},
			want: ("3.000000300"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.a.String(), func(t *testing.T) {
			if got := toDecimal(tt.a).StringFixed(9); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDecimal(%s) = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}

func TestToAmount(t *testing.T) {
	tests := []struct {
		i    decimal.Decimal
		want Amount
	}{
		{i: decimal.RequireFromString("0"),
			want: Amount{0, 0}},
		{i: decimal.RequireFromString("12345678.123456789"),
			want: Amount{12345678, 123456789}},
		{i: decimal.RequireFromString("-12345678.123456789"),
			want: Amount{-12345678, -123456789}},
		{i: decimal.RequireFromString("0.123456789012345"),
			want: Amount{0, 123456789}},
	}
	for _, tt := range tests {
		t.Run(tt.i.String(), func(t *testing.T) {
			if got := ToAmount(tt.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToAmount(%s) = %v, want %v", tt.i, got, tt.want)
			}
		})
	}
}

func Test_makeTrade(t *testing.T) {
	type args struct {
		p        *Portfolio
		action   grpcoin.TradeAction
		ticker   string
		quote    *grpcoin.Amount
		quantity *grpcoin.Amount
	}
	tests := []struct {
		name   string
		args   args
		code   codes.Code
		errMsg string
		want   *Portfolio
	}{
		{name: "no position initiated",
			args: args{p: &Portfolio{},
				action: grpcoin.TradeAction_SELL,
				ticker: "BTC"},
			code:   codes.InvalidArgument,
			errMsg: "position in portfolio"},
		{name: "insufficient cash",
			args: args{p: &Portfolio{
				CashUSD:   Amount{Units: 100, Nanos: 500_000_000},
				Positions: map[string]Amount{"BTC": {}},
			},
				action:   grpcoin.TradeAction_BUY,
				ticker:   "BTC",
				quote:    &grpcoin.Amount{Units: 200},
				quantity: &grpcoin.Amount{Units: 2},
			},
			code:   codes.InvalidArgument,
			errMsg: "insufficient cash after transaction (-299.5)"},
		{name: "insufficient positions",
			args: args{p: &Portfolio{
				CashUSD:   Amount{Units: 100},
				Positions: map[string]Amount{"BTC": {Units: 3, Nanos: 100_000_000}},
			},
				action:   grpcoin.TradeAction_SELL,
				ticker:   "BTC",
				quote:    &grpcoin.Amount{Units: 100},
				quantity: &grpcoin.Amount{Units: 3, Nanos: 200_000_000},
			},
			code:   codes.InvalidArgument,
			errMsg: "insufficient BTC positions (-0.1) after transaction"},
		{name: "successful buy",
			args: args{p: &Portfolio{
				CashUSD: Amount{Units: 100_000},
				Positions: map[string]Amount{
					"BTC": {Units: 3, Nanos: 100_000_000},
					"ETH": {Units: 3, Nanos: 100_000_000}},
			},
				action:   grpcoin.TradeAction_BUY,
				ticker:   "BTC",
				quote:    &grpcoin.Amount{Units: 50, Nanos: 300_000_000},
				quantity: &grpcoin.Amount{Units: 3},
			},
			code: codes.OK,
			want: &Portfolio{
				CashUSD: Amount{Units: 99_849, Nanos: 100_000_000},
				Positions: map[string]Amount{
					"BTC": {Units: 6, Nanos: 100_000_000},
					"ETH": {Units: 3, Nanos: 100_000_000}},
			},
		},
		{name: "successful sell",
			args: args{p: &Portfolio{
				CashUSD: Amount{Units: 100_000},
				Positions: map[string]Amount{
					"BTC": {Units: 3, Nanos: 100_000_000},
					"ETH": {Units: 3}},
			},
				action:   grpcoin.TradeAction_SELL,
				ticker:   "BTC",
				quote:    &grpcoin.Amount{Units: 50, Nanos: 300_000_000},
				quantity: &grpcoin.Amount{Units: 3},
			},
			code: codes.OK,
			want: &Portfolio{
				CashUSD: Amount{Units: 100_150, Nanos: 900_000_000},
				Positions: map[string]Amount{
					"BTC": {Units: 0, Nanos: 100_000_000},
					"ETH": {Units: 3}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := makeTrade(tt.args.p, tt.args.action, tt.args.ticker, tt.args.quote, tt.args.quantity)
			if status.Code(err) != tt.code {
				t.Errorf("makeTrade() error = %v, want code=%v", err, tt.code)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error doesn't contain %q, error: %v", tt.errMsg, err)
			}
			if err == nil {
				diff := cmp.Diff(tt.args.p, tt.want)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
