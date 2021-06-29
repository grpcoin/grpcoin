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
	"strings"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	ratelimiter2 "github.com/grpcoin/grpcoin/ratelimiter"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/grpcoin/grpcoin/apiserver/auth"
)

const (
	authenticatedRateLimitPerMinute   = 100
	unauthenticatedRateLimitPerMinute = 50
)

func rateLimitInterceptor(rl ratelimiter2.RateLimiter) grpc_auth.AuthFunc {
	return func(rpcCtx context.Context) (context.Context, error) {
		lg := ctxzap.Extract(rpcCtx).With(zap.String("facility", "rate"))
		u := auth.AuthInfoFromContext(rpcCtx)
		if u != nil {
			lg.Debug("rate check for user", zap.String("uid", u.DBKey()))
			key := "api_user__" + u.DBKey()
			return rpcCtx, rl.Hit(rpcCtx, key, authenticatedRateLimitPerMinute)
		}
		ip, err := findIP(rpcCtx)
		if err != nil {
			return rpcCtx, err
		} else if ip != "" {
			key := "api_ip__" + ip
			return rpcCtx, rl.Hit(rpcCtx, key, unauthenticatedRateLimitPerMinute)
		}
		lg.Warn("no ip or uid found in req ctx")
		return rpcCtx, nil
	}
}

// findIP extracts IP address from grpc request context.
// If no IP is found, returns empty string.
func findIP(rpcCtx context.Context) (string, error) {
	m, _ := metadata.FromIncomingContext(rpcCtx)
	if v := m.Get("x-real-ip"); len(v) > 0 {
		return v[0], nil
	}
	if v := m.Get("x-forwarded-for"); len(v) > 0 {
		ips := strings.Split(v[0], ",")
		return strings.TrimSpace(ips[0]), nil
	}
	peer, ok := peer.FromContext(rpcCtx)
	if !ok {
		return "", nil
	}
	ip, _, err := net.SplitHostPort(peer.Addr.String())
	return ip, err
}
