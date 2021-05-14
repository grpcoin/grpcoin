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
	"errors"
	"time"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tradingService struct {
	udb *userdb.UserDB
	tp  QuoteProvider
	tr  trace.Tracer

	grpcoin.UnimplementedPaperTradeServer
}

func (t *tradingService) Portfolio(ctx context.Context, req *grpcoin.PortfolioRequest) (*grpcoin.PortfolioResponse, error) {
	ctx, span := t.tr.Start(ctx, "portfolio read")
	defer span.End()
	user, ok := userdb.UserRecordFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "could not find user record in request context")
	}
	var pp []*grpcoin.PortfolioPosition
	for k, v := range user.Portfolio.Positions {
		pp = append(pp, &grpcoin.PortfolioPosition{
			Ticker: &grpcoin.PortfolioPosition_Ticker{Ticker: k},
			Amount: v.V(),
		})
	}
	return &grpcoin.PortfolioResponse{
		CashUsd:   user.Portfolio.CashUSD.V(),
		Positions: pp,
	}, nil
}

const (
	quoteDeadline          = time.Second * 2
	tradeExecutionDeadline = time.Second * 1
)

func (t *tradingService) Trade(ctx context.Context, req *grpcoin.TradeRequest) (*grpcoin.TradeResponse, error) {
	subCtx, s := t.tr.Start(ctx, "userctx")
	user, ok := userdb.UserRecordFromContext(subCtx)
	if !ok {
		return nil, status.Error(codes.Internal, "could not find user record in request context")
	}
	s.End()
	if err := validateTradeRequest(req); err != nil {
		return nil, err
	}

	// get a real-time market quote
	subCtx, s = t.tr.Start(ctx, "get quote")
	quoteCtx, cancel := context.WithTimeout(subCtx, quoteDeadline)
	defer cancel()
	quote, err := t.tp.GetQuote(quoteCtx, req.GetTicker().Ticker)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, status.Errorf(codes.Unavailable, "could not get real-time market quote for %s in %v",
			req.GetTicker().GetTicker(), quoteDeadline)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve a quote: %v", err)
	}
	s.End()

	// TODO add a timeout for tx to be executed
	subCtx, s = t.tr.Start(ctx, "execute trade")
	defer s.End()
	tradeCtx, cancel2 := context.WithTimeout(subCtx, tradeExecutionDeadline)
	defer cancel2()
	err = t.udb.Trade(tradeCtx, user.ID, req.GetTicker().Ticker, req.Action, quote, req.Quantity)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, status.Error(codes.Unavailable, "could not execute trade in a timely manner: %v")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute trade: %v", err)
	}

	return &grpcoin.TradeResponse{
		T:             timestamppb.Now(), // TODO read from tx
		Action:        req.Action,
		ExecutedPrice: quote,
		Quantity:      req.Quantity,
	}, nil
}

func validateTradeRequest(req *grpcoin.TradeRequest) error {
	if req.Action != grpcoin.TradeAction_BUY && req.Action != grpcoin.TradeAction_SELL {
		return status.Errorf(codes.InvalidArgument, "invalid trade action: %s", req.GetAction().Enum().String())
	}
	if req.GetQuantity() == nil || (req.GetQuantity().Nanos == 0 && req.GetQuantity().Units == 0) {
		return status.Error(codes.InvalidArgument, "quantity not specified")
	}
	if req.GetQuantity().Units < 0 {
		return status.Errorf(codes.InvalidArgument, "negative quantity units (%d)", req.GetQuantity().GetUnits())
	}
	if req.GetQuantity().Nanos < 0 {
		return status.Errorf(codes.InvalidArgument, "negative quantity nanos (%d)", req.GetQuantity().GetNanos())
	}
	if req.GetTicker().GetTicker() == "" {
		return status.Error(codes.InvalidArgument, "ticker not specified")
	}
	if req.GetTicker().GetTicker() != "BTC" {
		return status.Errorf(codes.InvalidArgument, "ticker '%s' not specified, only 'BTC' is supported", req.Ticker.Ticker)
	}
	return nil
}
