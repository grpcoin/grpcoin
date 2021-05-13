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
