package main

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/auth/github"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/firestoretestutil"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestPortfolio(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	fs := firestoretestutil.StartEmulator(t, context.TODO())
	au := auth.MockAuthenticator{
		F: func(c context.Context) (auth.AuthenticatedUser, error) {
			return &github.GitHubUser{ID: 1, Username: "abc"}, nil
		},
	}
	udb := &userdb.UserDB{DB: fs}
	lg, _ := zap.NewDevelopment()
	pt := &tradingService{udb: udb}
	srv := prepServer(context.TODO(), lg, au, udb, &accountService{cache: &AccountCache{cache: dummyRedis()}}, nil, pt)
	go srv.Serve(l)
	defer srv.Stop()
	defer l.Close()

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := grpcoin.NewPaperTradeClient(cc)

	resp, err := client.Portfolio(context.TODO(), &grpcoin.PortfolioRequest{})
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
