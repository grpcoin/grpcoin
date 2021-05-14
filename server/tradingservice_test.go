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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/auth/github"
	"github.com/grpcoin/grpcoin/server/firestoretestutil"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ QuoteProvider = &coinbaseQuoteProvider{}

type mockQuoteProvider struct {
	a   *grpcoin.Amount
	err error
}

func (m *mockQuoteProvider) GetQuote(ctx context.Context, ticker string) (*grpcoin.Amount, error) {
	return m.a, m.err
}

func TestPortfolio(t *testing.T) {
	fs := firestoretestutil.StartEmulator(t, context.TODO())
	udb := &userdb.UserDB{DB: fs, T: trace.NewNoopTracerProvider().Tracer("")}

	au := &github.GitHubUser{ID: 1, Username: "abc"}
	user, err := udb.EnsureAccountExists(context.TODO(), au)
	if err != nil {
		t.Fatal(err)
	}
	pt := &tradingService{udb: udb}

	ctx := auth.WithUser(context.Background(), au)
	ctx = userdb.WithUserRecord(ctx, user)

	resp, err := pt.Portfolio(ctx, &grpcoin.PortfolioRequest{})
	if err != nil {
		t.Fatal(err)
	}

	expected := &grpcoin.PortfolioResponse{
		CashUsd: &grpcoin.Amount{Units: 100_000, Nanos: 0},
		Positions: []*grpcoin.PortfolioPosition{
			{
				Ticker: &grpcoin.PortfolioPosition_Ticker{Ticker: "BTC"},
				Amount: &grpcoin.Amount{},
			},
		},
	}

	diff := cmp.Diff(resp, expected, cmpopts.IgnoreUnexported(
		grpcoin.PortfolioPosition{},
		grpcoin.PortfolioPosition_Ticker{},
		grpcoin.PortfolioResponse{},
		grpcoin.Amount{},
	))
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestTradeQuotePrices(t *testing.T) {
	fs := firestoretestutil.StartEmulator(t, context.TODO())
	udb := &userdb.UserDB{DB: fs, T: trace.NewNoopTracerProvider().Tracer("")}

	faultyQuote := &mockQuoteProvider{err: context.DeadlineExceeded}
	pt := &tradingService{udb: udb, tp: faultyQuote}
	au := &github.GitHubUser{ID: 1, Username: "abc"}
	user, err := udb.EnsureAccountExists(context.TODO(), au)
	if err != nil {
		t.Fatal(err)
	}

	ctx := auth.WithUser(context.Background(), au)
	ctx = userdb.WithUserRecord(ctx, user)

	_, err = pt.Trade(context.TODO(), &grpcoin.TradeRequest{})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected internal error when unauthenticated: %v", err)
	}

	_, err = pt.Trade(ctx, &grpcoin.TradeRequest{
		Action:   grpcoin.TradeAction_BUY,
		Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
		Quantity: &grpcoin.Amount{Units: 1},
	})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected unavailable error when quote cannot be recvd: %v", err)
	}

	pt.tp = &mockQuoteProvider{a: &grpcoin.Amount{Units: 50_000}}
	resp, err := pt.Trade(ctx, &grpcoin.TradeRequest{
		Action:   grpcoin.TradeAction_BUY,
		Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
		Quantity: &grpcoin.Amount{Units: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(resp.GetExecutedPrice(), &grpcoin.Amount{Units: 50_000},
		cmpopts.IgnoreUnexported(grpcoin.Amount{})); diff != "" {
		t.Fatal(diff)
	}
}

func Test_validateTradeRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *grpcoin.TradeRequest
		code codes.Code
	}{
		{
			name: "empty",
			req:  &grpcoin.TradeRequest{},
			code: codes.InvalidArgument,
		},
		{
			name: "empty ticker",
			req:  &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL},
			code: codes.InvalidArgument,
		},
		{
			name: "wrong ticker",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker: &grpcoin.TradeRequest_Ticker{Ticker: "XXX"}},
			code: codes.InvalidArgument,
		},
		{
			name: "missing quantity",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker: &grpcoin.TradeRequest_Ticker{Ticker: "BTC"}},
			code: codes.InvalidArgument,
		},
		{
			name: "zero quantity",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
				Quantity: &grpcoin.Amount{}},
			code: codes.InvalidArgument,
		},
		{
			name: "negative unit",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
				Quantity: &grpcoin.Amount{Units: -1},
			},
			code: codes.InvalidArgument,
		},
		{
			name: "negative nanos",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
				Quantity: &grpcoin.Amount{Nanos: -1},
			},
			code: codes.InvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTradeRequest(tt.req); status.Code(err) != tt.code {
				t.Errorf("validateTradeRequest() error = %v, wantErr %s", err, tt.code)
			}
		})
	}
}

func TestTrade(t *testing.T) {
	if testing.Short() {
		t.Skip("makes calls to coinbase")
	}
	fs := firestoretestutil.StartEmulator(t, context.TODO())
	udb := &userdb.UserDB{DB: fs, T: trace.NewNoopTracerProvider().Tracer("")}

	au := &github.GitHubUser{ID: 2, Username: "def"}
	user, err := udb.EnsureAccountExists(context.TODO(), au)
	if err != nil {
		t.Fatal(err)
	}
	ctx, stop := context.WithCancel(context.Background())
	t.Cleanup(func() { stop() })
	cb := &coinbaseQuoteProvider{}
	go cb.sync(ctx, "BTC")
	pt := &tradingService{udb: udb, tp: cb}

	ctx = auth.WithUser(ctx, au)
	ctx = userdb.WithUserRecord(ctx, user)

	// wait until we get a quote (it can take several seconds)
	_, err = cb.GetQuote(context.TODO(), "BTC")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := pt.Trade(ctx, &grpcoin.TradeRequest{
		Action:   grpcoin.TradeAction_BUY,
		Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
		Quantity: &grpcoin.Amount{Units: 1, Nanos: 500_000_000},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = resp
}
