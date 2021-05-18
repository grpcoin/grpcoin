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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/firestoretestutil"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testUser struct {
	id   string
	name string
}

func (t testUser) DBKey() string       { return t.id }
func (t testUser) DisplayName() string { return t.name }
func (t testUser) ProfileURL() string  { return "https://" + t.name }

func TestGetUser_notFound(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "foo"}

	u, ok, err := udb.Get(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("was not expecting to find user: %#v", u)
	}
}

func TestNewUser(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "foobar", name: "ab"}

	err := udb.Create(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}

	uv, ok, err := udb.Get(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not found created user")
	}

	expected := User{
		ID:          "foobar",
		DisplayName: "ab",
		ProfileURL:  "https://ab",
		Portfolio: Portfolio{CashUSD: Amount{Units: 100_000},
			Positions: map[string]Amount{
				"BTC": {Units: 0, Nanos: 0},
			}},
	}
	if diff := cmp.Diff(uv, expected,
		cmpopts.IgnoreFields(User{}, "CreatedAt")); diff != "" {
		t.Fatal(diff)
	}
}

func TestEnsureAccountExists(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "testuser", name: "abc"}

	u, err := udb.EnsureAccountExists(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == "" {
		t.Fatal("id should not be empty")
	}
	u2, err := udb.EnsureAccountExists(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(u, u2); diff != "" {
		t.Fatal(diff)
	}
}

func TestUserDB_Trade_OrderHistory(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "testuser", name: "abc"}
	if _, err := udb.EnsureAccountExists(ctx, tu); err != nil {
		t.Fatal(err)
	}

	// bad order
	if err := udb.Trade(ctx, tu.DBKey(), "BTC",
		grpcoin.TradeAction_SELL,
		&grpcoin.Amount{Units: 100},
		&grpcoin.Amount{Units: 100}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalidargument error: %v", err)
	}

	// several good orders
	if err := udb.Trade(ctx, tu.DBKey(), "BTC",
		grpcoin.TradeAction_BUY,
		&grpcoin.Amount{Units: 100},
		&grpcoin.Amount{Units: 25}); err != nil {
		t.Fatal(err)
	}
	if err := udb.Trade(ctx, tu.DBKey(), "BTC",
		grpcoin.TradeAction_SELL,
		&grpcoin.Amount{Units: 150},
		&grpcoin.Amount{Units: 10}); err != nil {
		t.Fatal(err)
	}
	if err := udb.Trade(ctx, tu.DBKey(), "BTC",
		grpcoin.TradeAction_SELL,
		&grpcoin.Amount{Units: 80},
		&grpcoin.Amount{Units: 10}); err != nil {
		t.Fatal(err)
	}

	// validate order history
	expectedOrders := []Order{
		{Ticker: "BTC", Action: grpcoin.TradeAction_BUY, Size: Amount{25, 0}, Price: Amount{100, 0}},
		{Ticker: "BTC", Action: grpcoin.TradeAction_SELL, Size: Amount{10, 0}, Price: Amount{150, 0}},
		{Ticker: "BTC", Action: grpcoin.TradeAction_SELL, Size: Amount{10, 0}, Price: Amount{80, 0}},
	}
	got, err := udb.UserOrderHistory(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(got, expectedOrders,
		cmpopts.IgnoreFields(Order{}, "Date"))
	if diff != "" {
		t.Fatal(diff)
	}
}
