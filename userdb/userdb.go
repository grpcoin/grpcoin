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
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/grpcoin/grpcoin/apiserver/firestoreutil"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/apiserver/auth"
)

type ctxUserRecordKey struct{}

const (
	fsUserCol      = "users"      // users collection
	fsTradesCol    = "orders"     // sub-collection for user
	fsValueHistCol = "valuations" // sub-collection for user's portfolio value over time

	maxTradeHistory         = 300  // keep latest N trade history records
	tradeHistoryRotateCheck = 0.01 // probability of checking & purging excess trade history records
)

var (
	// r is for probabilistic actions
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type User struct {
	ID          string
	DisplayName string
	ProfileURL  string
	CreatedAt   time.Time
	Portfolio   Portfolio
	TradeStats  struct {
		LastTrade  time.Time
		TradeCount int
	}
}

type Portfolio struct {
	CashUSD   Amount
	Positions map[string]Amount
}

type Amount struct {
	Units int64
	Nanos int32
}

func (a Amount) IsNegative() bool   { return a.Units < 0 || a.Nanos < 0 }
func (a Amount) IsZero() bool       { return a == Amount{} }
func (a Amount) V() *grpcoin.Amount { return &grpcoin.Amount{Units: a.Units, Nanos: a.Nanos} }
func (a Amount) F() decimal.Decimal { return toDecimal(a.V()) }
func (a Amount) Less(b Amount) bool {
	return a.Units < b.Units || (a.Units == b.Units && a.Nanos < b.Nanos)
}

// TradeRecord represents a trade user has made in the past.
type TradeRecord struct {
	Date   time.Time           `firestore:"date"`
	Ticker string              `firestore:"ticker"`
	Action grpcoin.TradeAction `firestore:"action"`
	Size   Amount              `firestore:"size"`
	Price  Amount              `firestore:"price"`
}

// ValuationHistory represents user's portfolio value at a particular time.
type ValuationHistory struct {
	Date  time.Time `firestore:"date"`
	Value Amount    `firestore:"value"`
}

type UserDB struct {
	DB    *firestore.Client
	Cache ProfileCache

	T trace.Tracer
}

func (u *UserDB) Create(ctx context.Context, au auth.AuthenticatedUser) error {
	newUser := User{ID: au.DBKey(),
		DisplayName: au.DisplayName(),
		ProfileURL:  au.ProfileURL(),
		CreatedAt:   time.Now(),
	}
	setupGamePortfolio(&newUser)
	_, err := u.DB.Collection(fsUserCol).Doc(au.DBKey()).Create(ctx, newUser)
	if err != nil {
		return err
	}

	return u.SetUserValuationHistory(ctx, newUser.ID, ValuationHistory{
		Date:  time.Now().UTC(),
		Value: defaultStartingCash})
}

func (u *UserDB) Get(ctx context.Context, userID string) (User, bool, error) {
	doc, err := u.DB.Collection(fsUserCol).Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return User{}, false, nil
		}
		return User{}, false, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}
	var uv User
	if err := doc.DataTo(&uv); err != nil {
		return User{}, false, fmt.Errorf("failed to unpack user record %q: %w", userID, err)
	}
	return uv, true, nil
}

func (u *UserDB) GetAll(ctx context.Context) ([]User, error) {
	var out []User
	iter := u.DB.Collection(fsUserCol).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var u User
		if err := doc.DataTo(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func (u *UserDB) EnsureAccountExists(ctx context.Context, au auth.AuthenticatedUser) (User, error) {
	ctx, s := u.T.Start(ctx, "ensure user account")
	defer s.End()
	user, exists, err := u.Get(ctx, au.DBKey())
	if exists {
		return user, nil
	} else if err != nil {
		return User{}, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}
	err = u.Create(ctx, au)
	if err != nil {
		return User{}, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	user, _, err = u.Get(ctx, au.DBKey())
	if err != nil {
		return User{}, status.Errorf(codes.Internal, "failed to query new user: %v", err)
	}
	return user, err
}

// EnsureAccountExistsInterceptor creates an account for the authenticated
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

func (u *UserDB) Trade(ctx context.Context, uid string, ticker string, action grpcoin.TradeAction,
	quote, quantity *grpcoin.Amount) (Portfolio, error) {
	subCtx, s := u.T.Start(ctx, "trade tx")
	ref := u.DB.Collection("users").Doc(uid)
	var resultingPortfolio Portfolio
	err := u.DB.RunTransaction(subCtx, func(ctx context.Context, tx *firestore.Transaction) error {
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
		u.TradeStats.LastTrade = time.Now()
		u.TradeStats.TradeCount++
		resultingPortfolio = u.Portfolio
		return tx.Set(ref, u)
	}, firestore.MaxAttempts(1))
	s.End()

	if err != nil {
		return resultingPortfolio, err
	}

	subCtx, s = u.T.Start(ctx, "log trade")
	err = u.recordTradeHistory(subCtx, uid, time.Now().UTC(), ticker, action,
		ToAmount(toDecimal(quantity)), ToAmount(toDecimal(quote)))
	if err != nil {
		s.RecordError(err)
		ctxzap.Extract(ctx).Warn("failed to record trade history", zap.Error(err))
	}
	s.End()

	subCtx, s = u.T.Start(ctx, "invalidate trade history cache")
	if err := u.Cache.InvalidateTrades(subCtx, uid); err != nil {
		s.RecordError(err)
		ctxzap.Extract(ctx).Warn("failed to invalidate trade history cache", zap.Error(err))
	}

	// probabilistically delete unnecessary trade history records
	if r.Float64() < tradeHistoryRotateCheck {
		subCtx, s := u.T.Start(ctx, "rotate trade history")
		defer s.End()
		if err := u.RotateTradeHistory(subCtx, uid, maxTradeHistory); err != nil {
			s.RecordError(err)
			ctxzap.Extract(ctx).Warn("failed to rotate trade history", zap.Error(err))
		}
	}
	return resultingPortfolio, nil // do not block trades on trade history bookkeeping
}

func (u *UserDB) recordTradeHistory(ctx context.Context, uid string,
	t time.Time, ticker string, action grpcoin.TradeAction, size, price Amount) error {
	id := t.Format(time.RFC3339Nano)
	_, err := u.DB.Collection(fsUserCol).Doc(uid).Collection(fsTradesCol).Doc(id).Create(ctx, TradeRecord{
		Date:   t,
		Ticker: ticker,
		Action: action,
		Size:   size,
		Price:  price,
	})
	return err
}

func (u *UserDB) UserTrades(ctx context.Context, uid string) ([]TradeRecord, error) {
	ctx, s := u.T.Start(ctx, "trade history")
	defer s.End()

	if v, ok, err := u.Cache.GetTrades(ctx, uid); err != nil {
		return nil, fmt.Errorf("failed to query trade history cache: %v", err)
	} else if ok {
		return v, nil
	}

	var out []TradeRecord
	iter := u.DB.Collection(fsUserCol).Doc(uid).Collection(fsTradesCol).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			s.RecordError(err)
			return nil, err
		}
		var v TradeRecord
		if err := doc.DataTo(&v); err != nil {
			s.RecordError(err)
			return nil, err
		}
		out = append(out, v)
	}

	if err := u.Cache.SaveTrades(ctx, uid, out); err != nil {
		ctxzap.Extract(ctx).Warn("failed to save trade history to cache", zap.String("uid", uid))
	}
	return out, nil
}

func (u *UserDB) RotateTradeHistory(ctx context.Context, uid string, maxHist int) error {
	it := u.DB.Collection(fsUserCol).Doc(uid).Collection(fsTradesCol).
		OrderBy("date", firestore.Desc).Offset(maxHist).Documents(ctx)
	return firestoreutil.BatchDeleteAll(ctx, u.DB, it)
}

func (u *UserDB) UserValuationHistory(ctx context.Context, uid string) ([]ValuationHistory, error) {
	// TODO implement caching around this with a hourly key and precise ttl.
	ctx, s := u.T.Start(ctx, "user valuation history")
	defer s.End()

	if v, ok, err := u.Cache.GetValuation(ctx, uid, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to retrieve valuation history: %v", err)
	} else if ok {
		return v, nil
	}

	var out []ValuationHistory
	iter := u.DB.Collection(fsUserCol).Doc(uid).Collection(fsValueHistCol).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			s.RecordError(err)
			return nil, err
		}
		var v ValuationHistory
		if err := doc.DataTo(&v); err != nil {
			s.RecordError(err)
			return nil, err
		}
		out = append(out, v)
	}
	if err := u.Cache.SaveValuation(ctx, uid, time.Now(), out); err != nil {
		ctxzap.Extract(ctx).Warn("failed to save portfolio valuation history", zap.String("uid", uid), zap.Int("size", len(out)))
	}
	return out, nil
}

func canonicalizeValuationHistoryDBKey(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func (u *UserDB) SetUserValuationHistory(ctx context.Context, uid string, v ValuationHistory) error {
	_, err := u.DB.Collection(fsUserCol).Doc(uid).Collection(fsValueHistCol).
		Doc(canonicalizeValuationHistoryDBKey(v.Date)).Create(ctx, v)
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return err
	}
	return nil
}

func (u *UserDB) RotateUserValuationHistory(ctx context.Context, uid string, deleteBefore time.Time) error {
	// TODO create an index for users/*/date ASC
	it := u.DB.Collection(fsUserCol).Doc(uid).Collection(fsValueHistCol).
		Where("date", "<", deleteBefore).Documents(ctx)
	return firestoreutil.BatchDeleteAll(ctx, u.DB, it)
}

func UserRecordFromContext(ctx context.Context) (User, bool) {
	v := ctx.Value(ctxUserRecordKey{})
	vv, ok := v.(User)
	return vv, ok
}

func WithUserRecord(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxUserRecordKey{}, u)
}
