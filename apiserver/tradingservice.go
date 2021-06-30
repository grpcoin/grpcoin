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
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/userdb"
)

type tradingService struct {
	udb              *userdb.UserDB
	quoteProvider    realtimequote.QuoteProvider
	tracer           trace.Tracer
	supportedTickers []string

	grpcoin.UnimplementedPaperTradeServer
}

func toPortfolioPositions(pos map[string]userdb.Amount) []*grpcoin.PortfolioPosition {
	var pp []*grpcoin.PortfolioPosition
	for k, v := range pos {
		pp = append(pp, &grpcoin.PortfolioPosition{
			Currency: &grpcoin.Currency{Symbol: k},
			Amount:   v.V(),
		})
	}
	return pp
}

func (t *tradingService) Portfolio(ctx context.Context, req *grpcoin.PortfolioRequest) (*grpcoin.PortfolioResponse, error) {
	ctx, span := t.tracer.Start(ctx, "portfolio read")
	defer span.End()
	user, ok := userdb.UserRecordFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "could not find user record in request context")
	}
	return &grpcoin.PortfolioResponse{
		CashUsd:   user.Portfolio.CashUSD.V(),
		Positions: toPortfolioPositions(user.Portfolio.Positions),
	}, nil
}

const (
	quoteDeadline          = time.Second * 2
	tradeExecutionDeadline = time.Second * 1
)

func (t *tradingService) Trade(ctx context.Context, req *grpcoin.TradeRequest) (*grpcoin.TradeResponse, error) {
	user, ok := userdb.UserRecordFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "could not find user record in request context")
	}
	if err := validateTradeRequest(req, t.supportedTickers); err != nil {
		return nil, err
	}
	product := req.GetCurrency().GetSymbol()

	// get a real-time market quote
	// TODO use oteltrace.WithAttributes for sub-spans
	subCtx, s := t.tracer.Start(ctx, "get quote")
	quoteCtx, cancel := context.WithTimeout(subCtx, quoteDeadline)
	defer cancel()
	quote, err := t.quoteProvider.GetQuote(quoteCtx, product)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, status.Errorf(codes.Unavailable, "could not get real-time market quote for %s in %v",
			product, quoteDeadline)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve a quote: %v", err)
	}
	s.End()

	// TODO add a timeout for tx to be executed
	subCtx, s = t.tracer.Start(ctx, "execute trade")
	defer s.End()
	tradeCtx, cancel2 := context.WithTimeout(subCtx, tradeExecutionDeadline)
	defer cancel2()
	newPortfolio, err := t.udb.Trade(tradeCtx, user.ID, product, req.Action, quote, req.Quantity)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, status.Errorf(codes.Unavailable, "could not execute trade in a timely manner: %v", err)
	} else if status.Code(err) == codes.InvalidArgument {
		return nil, err
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute trade: %v", err)
	}

	return &grpcoin.TradeResponse{
		T:             timestamppb.Now(), // TODO read from tx
		Action:        req.Action,
		ExecutedPrice: quote,
		Currency:      &grpcoin.Currency{Symbol: product},
		Quantity:      req.Quantity,
		ResultingPortfolio: &grpcoin.TradeResponse_Portfolio{
			RemainingCash: newPortfolio.CashUSD.V(),
			Positions:     toPortfolioPositions(newPortfolio.Positions),
		},
	}, nil
}

func validateTradeRequest(req *grpcoin.TradeRequest, supportedTickers []string) error {
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
	if req.GetCurrency().GetSymbol() == "" {
		return status.Error(codes.InvalidArgument, "ticker not specified")
	}
	if !realtimequote.IsSupported(supportedTickers, req.GetCurrency().GetSymbol()) {
		return status.Errorf(codes.InvalidArgument, "ticker '%s' is not supported, must be [%s]", req.GetCurrency().GetSymbol(),
			strings.Join(supportedTickers, ", "))
	}
	return nil
}
