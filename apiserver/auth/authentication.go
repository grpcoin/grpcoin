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

package auth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxAuthUserInfo struct{} // stores authenticated user info (e.g. github profile)

type AuthenticatedUser interface {
	DBKey() string
	DisplayName() string
	ProfileURL() string
}

type Authenticator interface {
	Authenticate(rpcCtx context.Context) (AuthenticatedUser, error)
}

func AuthenticatingInterceptor(a Authenticator) grpc_auth.AuthFunc {
	return func(rpcCtx context.Context) (context.Context, error) {
		user, err := a.Authenticate(rpcCtx)
		if err != nil {
			return rpcCtx, err
		}

		if span := trace.SpanFromContext(rpcCtx); span != nil {
			span.SetAttributes(attribute.String("user.id", user.DBKey()))
		}
		ctx := WithUser(rpcCtx, user)
		return ctx, nil
	}
}

// AuthInfoFromContext extracts authenticated user info from the ctx.
func AuthInfoFromContext(ctx context.Context) AuthenticatedUser {
	v := ctx.Value(ctxAuthUserInfo{})
	if v == nil {
		return nil
	}
	return v.(AuthenticatedUser)
}

// WithUser adds authenticated user into to ctx.
func WithUser(ctx context.Context, a AuthenticatedUser) context.Context {
	return context.WithValue(ctx, ctxAuthUserInfo{}, a)
}

type MockAuthenticator struct {
	F func(context.Context) (AuthenticatedUser, error)
}

func (m MockAuthenticator) Authenticate(ctx context.Context) (AuthenticatedUser, error) {
	return m.F(ctx)
}
