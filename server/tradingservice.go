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

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/userdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tradingService struct {
	udb *userdb.UserDB

	grpcoin.UnimplementedPaperTradeServer
}

func (t *tradingService) Portfolio(ctx context.Context, req *grpcoin.PortfolioRequest) (*grpcoin.PortfolioResponse, error) {
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

func (t *tradingService) Trade(ctx context.Context, req *grpcoin.TradeRequest) (*grpcoin.TradeResponse, error) {
	user, ok := userdb.UserRecordFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "could not find user record in request context")
	}
	if req.Action != grpcoin.TradeAction_BUY && req.Action != grpcoin.TradeAction_SELL {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trade action: %s", req.GetAction().Enum().String())
	}
	if req.Quantity.Units < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "negative quantity units (%d)", req.GetQuantity().GetUnits())
	}
	if req.Quantity.Nanos < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "negative quantity nanos (%d)", req.GetQuantity().GetNanos())
	}
	if req.GetTicker().GetTicker() == "" {
		return nil, status.Error(codes.InvalidArgument, "ticker not specified")
	}
	if req.GetTicker().GetTicker() != "BTC" {
		return nil, status.Errorf(codes.InvalidArgument, "ticker '%s' not specified, only 'BTC' is supported", req.Ticker.Ticker)
	}
	_ = user
	// TODO execute trade on UserDB.
	return &grpcoin.TradeResponse{
		T:             timestamppb.Now(), // TODO read from tx
		Action:        req.Action,
		ExecutedPrice: &grpcoin.Amount{}, // TODO plumb ticker service here
		Quantity:      req.Quantity,
	}, nil
}
