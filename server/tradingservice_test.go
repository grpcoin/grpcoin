package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/auth/github"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/firestoretestutil"
	"github.com/grpcoin/grpcoin/server/userdb"
)

func TestPortfolio(t *testing.T) {
	fs := firestoretestutil.StartEmulator(t, context.TODO())
	udb := &userdb.UserDB{DB: fs}

	ctx := auth.WithUser(context.Background(), &github.GitHubUser{ID: 1, Username: "abc"})
	pt := &tradingService{udb: udb}

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
