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

package userdb

import (
	"context"
	"fmt"
	"time"

	firestore "cloud.google.com/go/firestore"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/auth"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ctxUserRecordKey struct{}

const (
	fsUserCol = "users"
)

type User struct {
	ID          string
	DisplayName string
	ProfileURL  string
	CreatedAt   time.Time
	Portfolio   Portfolio
}

type Portfolio struct {
	CashUSD   Amount
	Positions map[string]Amount
}

type Amount struct {
	Units int64
	Nanos int32
}

func (a Amount) V() *grpcoin.Amount { return &grpcoin.Amount{Units: a.Units, Nanos: a.Nanos} }

type UserDB struct {
	DB *firestore.Client
	T  trace.Tracer
}

func (u *UserDB) Create(ctx context.Context, au auth.AuthenticatedUser) error {
	ctx, s := u.T.Start(ctx, "firestore.create")
	defer s.End()
	newUser := User{ID: au.DBKey(),
		DisplayName: au.DisplayName(),
		ProfileURL:  au.ProfileURL(),
		CreatedAt:   time.Now(),
	}
	setupGamePortfolio(&newUser)
	_, err := u.DB.Collection(fsUserCol).Doc(au.DBKey()).Create(ctx, newUser)
	return err
}

func (u *UserDB) Get(ctx context.Context, au auth.AuthenticatedUser) (User, bool, error) {
	ctx, s := u.T.Start(ctx, "firestore.get")
	defer s.End()
	doc, err := u.DB.Collection(fsUserCol).Doc(au.DBKey()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return User{}, false, nil
		}
		return User{}, false, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}
	var uv User
	if err := doc.DataTo(&uv); err != nil {
		return User{}, false, fmt.Errorf("failed to unpack user record %q: %w", au.DBKey(), err)
	}
	return uv, true, nil
}

func (u *UserDB) EnsureAccountExists(ctx context.Context, au auth.AuthenticatedUser) (User, error) {
	ctx, s := u.T.Start(ctx, "ensure acct")
	defer s.End()
	user, _, err := u.Get(ctx, au)
	if err == nil {
		return user, err
	} else if status.Code(err) != codes.NotFound {
		return User{}, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}
	err = u.Create(ctx, au)
	if err != nil {
		return User{}, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	user, _, err = u.Get(ctx, au)
	if err != nil {
		return User{}, status.Errorf(codes.Internal, "failed to query new user: %v", err)
	}
	return user, err
}

// ensureAccountExistsInterceptor creates an account for the authenticated
// client (or retrieves it) and augments the ctx with the user's db record.
func (u *UserDB) EnsureAccountExistsInterceptor() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		authenticatedUser := auth.AuthInfoFromContext(ctx)
		if authenticatedUser == nil {
			return ctx, status.Error(codes.Internal, "req ctx did not have user info")
		}
		v, ok := authenticatedUser.(auth.AuthenticatedUser)
		if !ok {
			return ctx, status.Errorf(codes.Internal, "unknown authed user type %T", authenticatedUser)
		}
		uv, err := u.EnsureAccountExists(ctx, v)
		if err != nil {
			return ctx, status.Errorf(codes.Internal, "failed to ensure user account: %v", err)
		}
		return WithUserRecord(ctx, uv), nil
	}
}

func (u *UserDB) Trade(ctx context.Context, uid string, ticker string, action grpcoin.TradeAction, quote, quantity *grpcoin.Amount) error {
	ref := u.DB.Collection("users").Doc(uid)
	err := u.DB.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return fmt.Errorf("failed to read user record for tx: %w", err)
		}
		var u User
		if err := doc.DataTo(&u); err != nil {
			return fmt.Errorf("failed to unpack user record into struct: %w", err)
		}
		if err := makeTrade(&u.Portfolio, action, ticker, quote, quantity); err != nil {
			return err
		}
		return tx.Set(ref, u)
	}, firestore.MaxAttempts(1))
	return err
}

func UserRecordFromContext(ctx context.Context) (User, bool) {
	v := ctx.Value(ctxUserRecordKey{})
	vv, ok := v.(User)
	return vv, ok
}

func WithUserRecord(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxUserRecordKey{}, u)
}
