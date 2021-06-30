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
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	prod  = `api.grpco.in:443`
	local = `localhost:8080`
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Ltime)
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("create a permissionless Personal Access Token on GitHub https://github.com/settings/tokens and set it to TOKEN environment variable")
	}
	ctx := context.Background()

	var conn *grpc.ClientConn
	if os.Getenv("LOCAL") == "" {
		c, err := grpc.DialContext(ctx, prod, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
		if err != nil {
			log.Fatal(err)
		}
		conn = c
	} else {
		c, err := grpc.DialContext(ctx, local, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		conn = c
	}

	// try adding token to outgoing request
	authCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	resp, err := grpcoin.NewAccountClient(conn).TestAuth(authCtx, &grpcoin.TestAuthRequest{})
	if err != nil {
		log.Fatalf("authentication failed: %v", err)
	}
	log.Printf("you are user %s", resp.GetUserId())

	// retrieve portfolio
	portfolio, err := grpcoin.NewPaperTradeClient(conn).Portfolio(authCtx, &grpcoin.PortfolioRequest{})
	if err != nil {
		log.Fatalf("portfolio request failed: %v", err)
	}
	log.Printf("cash position: USD %s", fmtAmount(portfolio.CashUsd))
	for _, p := range portfolio.Positions {
		log.Printf("-> coin position: %s (%s)", p.Currency.Symbol, fmtAmount(p.Amount))
	}

	// buy 0.05 btc
	order, err := grpcoin.NewPaperTradeClient(conn).Trade(authCtx, &grpcoin.TradeRequest{
		Action:   grpcoin.TradeAction_BUY,
		Currency: &grpcoin.Currency{Symbol: "BTC"},
		Quantity: &grpcoin.Amount{Units: 0, Nanos: 50_000_000},
	})
	if err != nil {
		log.Fatalf("trade order failed: %v", err)
	}
	log.Printf("ORDER EXECUTED: %s [%s] %s at USD[%s] (cash remaining: %s)", order.Action,
		fmtAmount(order.Quantity), order.Currency.Symbol, fmtAmount(order.ExecutedPrice),
		fmtAmount(order.ResultingPortfolio.RemainingCash))

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	for ctx.Err() == nil {
		log.Printf("\n\nconnecting to stream real-time ETH quotes, hit Ctrl-C to quit anytime")
		stream, err := grpcoin.NewTickerInfoClient(conn).Watch(ctx,
			&grpcoin.TickerWatchRequest{Currency: &grpcoin.Currency{Symbol: "ETH"}})
		if err != nil {
			log.Fatal(err)
		}
		for ctx.Err() == nil {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				if v, ok := status.FromError(err); ok && v.Code() == codes.Canceled {
					break
				}
				log.Fatalf("unexpected: %v", err)
			}
			log.Printf("[server:%s] --  %d.%d",
				msg.T.AsTime().Format(time.RFC3339Nano),
				msg.Price.GetUnits(),
				msg.Price.GetNanos())
		}
		log.Printf("disconnected")
		time.Sleep(time.Second)
	}
}

func fmtAmount(a *grpcoin.Amount) string {
	s := fmt.Sprintf("%d.%09d", a.Units, a.Nanos)
	s = strings.TrimSuffix(s, "0000000")
	return s
}
