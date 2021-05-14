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

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountServiceOpts struct {
	redisIP string
}

type accountService struct {
	cache *AccountCache
	udb   *userdb.UserDB

	grpcoin.UnimplementedAccountServer
}

func (s *accountService) createAccount(ctx context.Context) {
	panic("unimplemented")
}

func (s *accountService) TestAuth(ctx context.Context, req *grpcoin.TestAuthRequest) (*grpcoin.TestAuthResponse, error) {
	log := ctxzap.Extract(ctx)
	v := auth.AuthInfoFromContext(ctx)
	if v == nil {
		return nil, status.Error(codes.Internal, "request arrived without a token")
	}
	log.Debug("authenticated user", zap.String("name", v.DisplayName()))
	u, ok := userdb.UserRecordFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "no user record in request context")
	}
	return &grpcoin.TestAuthResponse{UserId: u.ID}, nil
}
