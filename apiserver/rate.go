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
	"net"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/grpcoin/grpcoin/apiserver/auth"
	"github.com/grpcoin/grpcoin/apiserver/ratelimiter"
)

const (
	authenticatedRateLimitPerMinute   = 100
	unauthenticatedRateLimitPerMinute = 50
)

func rateLimitInterceptor(rl ratelimiter.RateLimiter) grpc_auth.AuthFunc {
	return func(rpcCtx context.Context) (context.Context, error) {
		lg := ctxzap.Extract(rpcCtx).With(zap.String("facility", "rate"))
		u := auth.AuthInfoFromContext(rpcCtx)
		if u != nil {
			lg.Debug("rate check for user", zap.String("uid", u.DBKey()))
			return rpcCtx, rl.Hit(rpcCtx, u.DBKey(), authenticatedRateLimitPerMinute)
		}
		m, ok := metadata.FromIncomingContext(rpcCtx)
		if ok && len(m.Get("x-forwarded-for")) > 0 {
			ip := m.Get("x-forwarded-for")[0]
			lg.Debug("rate check for ip", zap.String("ip", ip))
			return rpcCtx, rl.Hit(rpcCtx, ip, unauthenticatedRateLimitPerMinute)
		}
		peer, ok := peer.FromContext(rpcCtx)
		if ok {
			ip, _, err := net.SplitHostPort(peer.Addr.String())
			if err == nil {
				lg.Debug("rate check for ip", zap.String("ip", ip))
				return rpcCtx, rl.Hit(rpcCtx, ip, unauthenticatedRateLimitPerMinute)
			} else {
				lg.Warn("failed to parse host/port", zap.String("v", peer.Addr.String()))
			}
		}
		lg.Warn("no ip or uid found in req ctx")
		return rpcCtx, nil
	}
}

