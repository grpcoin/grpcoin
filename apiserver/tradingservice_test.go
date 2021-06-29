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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/realtimequote"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/apiserver/auth"
	"github.com/grpcoin/grpcoin/apiserver/auth/github"
	"github.com/grpcoin/grpcoin/apiserver/firestoreutil"
	"github.com/grpcoin/grpcoin/userdb"
)

type mockQuoteStream struct {
	product string
	price   *grpcoin.Amount
	n       int
}

func (m mockQuoteStream) Watch(ctx context.Context, _ ...string) (<-chan realtimequote.Quote, error) {
	ch := make(chan realtimequote.Quote)
	go func() {
		tick := time.NewTicker(time.Millisecond * 10)
		defer tick.Stop()
		for i := 0; i < m.n; i++ {
			select {
			case t := <-tick.C:
				ch <- realtimequote.Quote{
					Product: m.product,
					Price:   m.price,
					Time:    t}
			case <-ctx.Done():
			}
		}
	}()
	return ch, ctx.Err()
}

type mockQuoteProvider struct {
	a   *grpcoin.Amount
	err error
}

func (m *mockQuoteProvider) GetQuote(_ context.Context, _ string) (*grpcoin.Amount, error) {
	return m.a, m.err
}

func TestPortfolio(t *testing.T) {
	fs := firestoreutil.StartTestEmulator(t, context.TODO())
	tp := trace.NewNoopTracerProvider().Tracer("")
	udb := &userdb.UserDB{DB: fs, T: tp}

	au := &github.GitHubUser{ID: 1, Username: "abc"}
	user, err := udb.EnsureAccountExists(context.TODO(), au)
	if err != nil {
		t.Fatal(err)
	}
	pt := &tradingService{udb: udb, tracer: tp, supportedTickers: []string{"BTC"}}
	ctx := auth.WithUser(context.Background(), au)
	ctx = userdb.WithUserRecord(ctx, user)

	resp, err := pt.Portfolio(ctx, &grpcoin.PortfolioRequest{})
	if err != nil {
		t.Fatal(err)
	}

	expected := &grpcoin.PortfolioResponse{
		CashUsd:   &grpcoin.Amount{Units: 100_000, Nanos: 0},
		Positions: nil,
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
	fs := firestoreutil.StartTestEmulator(t, context.TODO())
	tr := trace.NewNoopTracerProvider().Tracer("")
	udb := &userdb.UserDB{DB: fs, T: tr}

	faultyQuote := &mockQuoteProvider{err: context.DeadlineExceeded}
	pt := &tradingService{udb: udb, quoteProvider: faultyQuote, tracer: tr, supportedTickers: []string{"BTC"}}
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

	pt.quoteProvider = &mockQuoteProvider{a: &grpcoin.Amount{Units: 50_000}}
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
		name             string
		req              *grpcoin.TradeRequest
		supportedTickers []string
		code             codes.Code
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
			supportedTickers: []string{},
			code:             codes.InvalidArgument,
		},
		{
			name: "missing quantity",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker: &grpcoin.TradeRequest_Ticker{Ticker: "BTC"}},
			supportedTickers: []string{"ABC", "BTC"},
			code:             codes.InvalidArgument,
		},
		{
			name: "zero quantity",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
				Quantity: &grpcoin.Amount{}},
			supportedTickers: []string{"ABC", "BTC"},
			code:             codes.InvalidArgument,
		},
		{
			name: "negative unit",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
				Quantity: &grpcoin.Amount{Units: -1},
			},
			supportedTickers: []string{"ABC", "BTC"},
			code:             codes.InvalidArgument,
		},
		{
			name: "negative nanos",
			req: &grpcoin.TradeRequest{Action: grpcoin.TradeAction_SELL,
				Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
				Quantity: &grpcoin.Amount{Nanos: -1},
			},
			supportedTickers: []string{"ABC", "BTC"},
			code:             codes.InvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTradeRequest(tt.req, tt.supportedTickers); status.Code(err) != tt.code {
				t.Errorf("validateTradeRequest() error = %v, wantErr %s", err, tt.code)
			}
		})
	}
}

func TestTrade(t *testing.T) {
	tp := trace.NewNoopTracerProvider().Tracer("")
	fs := firestoreutil.StartTestEmulator(t, context.TODO())
	udb := &userdb.UserDB{DB: fs, T: tp}

	au := &github.GitHubUser{ID: 2, Username: "def"}
	user, err := udb.EnsureAccountExists(context.TODO(), au)
	if err != nil {
		t.Fatal(err)
	}
	ctx, stop := context.WithCancel(context.Background())
	t.Cleanup(func() { stop() })
	qp := &mockQuoteProvider{a: &grpcoin.Amount{Units: 30_000}}
	pt := &tradingService{udb: udb, quoteProvider: qp, tracer: tp, supportedTickers: []string{"BTC"}}

	ctx = auth.WithUser(ctx, au)
	ctx = userdb.WithUserRecord(ctx, user)

	// insufficient funds
	_, err = pt.Trade(ctx, &grpcoin.TradeRequest{
		Action:   grpcoin.TradeAction_BUY,
		Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
		Quantity: &grpcoin.Amount{Units: 100_000_000},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for wrong trade request (insufficient funds): %v", err)
	}

	// good trade
	resp, err := pt.Trade(ctx, &grpcoin.TradeRequest{
		Action:   grpcoin.TradeAction_BUY,
		Ticker:   &grpcoin.TradeRequest_Ticker{Ticker: "BTC"},
		Quantity: &grpcoin.Amount{Units: 1, Nanos: 500_000_000},
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := &grpcoin.TradeResponse{
		T:             nil,
		Action:        grpcoin.TradeAction_BUY,
		Ticker:        &grpcoin.TradeResponse_Ticker{Symbol: "BTC"},
		Quantity:      &grpcoin.Amount{Units: 1, Nanos: 500_000_000},
		ExecutedPrice: &grpcoin.Amount{Units: 30_000},
		ResultingPortfolio: &grpcoin.TradeResponse_Portfolio{
			RemainingCash: &grpcoin.Amount{Units: 55_000, Nanos: 0},
			Positions: []*grpcoin.PortfolioPosition{
				{Ticker: &grpcoin.PortfolioPosition_Ticker{Ticker: "BTC"},
					Amount: &grpcoin.Amount{Units: 1, Nanos: 500_000_000}},
			},
		},
	}
	if diff := cmp.Diff(*expected, *resp,
		cmpopts.IgnoreUnexported(grpcoin.TradeResponse{},
			grpcoin.TradeResponse_Ticker{},
			grpcoin.Amount{},
			grpcoin.PortfolioPosition{},
			grpcoin.PortfolioPosition_Ticker{},
			grpcoin.TradeResponse_Portfolio{}),
		cmpopts.IgnoreFields(grpcoin.TradeResponse{}, "T"),
	); diff != "" {
		t.Fatalf(diff)
	}

	u, _, err := udb.Get(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if expected, trades := 1, u.TradeStats.TradeCount; trades != expected {
		t.Fatalf("expected trade count=%d, got=%d", expected, trades)
	}
	if s := time.Since(u.TradeStats.LastTrade); s > time.Second {
		t.Fatalf("more than 1s (%d) since last trade date", s)
	}
}
