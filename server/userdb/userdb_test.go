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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/firestoreutil"
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
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}

	u, ok, err := udb.Get(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("was not expecting to find user: %#v", u)
	}
}

func TestNewUser(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "foobar", name: "ab"}

	err := udb.Create(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}

	uv, ok, err := udb.Get(ctx, "foobar")
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

	vals, err := udb.UserValuationHistory(ctx, "foobar")
	if len(vals) != 1 {
		t.Fatalf("new user should have 1 valuation record: %#v", vals)
	}
}

func TestEnsureAccountExists(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
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

func TestValuationHistory(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "testuser", name: "abc"}
	u, err := udb.EnsureAccountExists(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	ti := time.Date(2050, 03, 12, 0, 0, 0, 0, time.UTC)
	v1 := ValuationHistory{Date: ti, Value: Amount{Units: 5555}}
	v2 := ValuationHistory{Date: ti, Value: Amount{Units: 6666}}
	if err := udb.SetUserValuationHistory(ctx, u.ID, v1); err != nil {
		t.Fatal(err)
	}
	if err := udb.SetUserValuationHistory(ctx, u.ID, v2); err != nil {
		t.Fatal(err)
	}

	v, err := udb.UserValuationHistory(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) > 1 {
		v = v[1:] // remove signup record
	}
	expected := []ValuationHistory{v1}
	diff := cmp.Diff(expected, v)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestRotateUserValuationHistory(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "testuser", name: "abc"}
	u, err := udb.EnsureAccountExists(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}
	d := time.Date(2050, time.April, 23, 0, 0, 0, 0, time.UTC)
	for i := 1; i <= 20; i++ {
		dv := d.Add(time.Hour * time.Duration(i))
		if err := udb.SetUserValuationHistory(ctx, u.ID,
			ValuationHistory{Date: dv}); err != nil {
			t.Fatal(err)
		}
	}

	if err := udb.RotateUserValuationHistory(ctx, u.ID, d.Add(time.Hour*15)); err != nil {
		t.Fatal(err)
	}

	v, err := udb.UserValuationHistory(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	expected := []ValuationHistory{
		{Date: d.Add(time.Hour * 15)},
		{Date: d.Add(time.Hour * 16)},
		{Date: d.Add(time.Hour * 17)},
		{Date: d.Add(time.Hour * 18)},
		{Date: d.Add(time.Hour * 19)},
		{Date: d.Add(time.Hour * 20)},
	}
	diff := cmp.Diff(expected, v)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestUserDB_Trade_OrderHistory(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
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
	got, err := udb.UserOrderHistory(ctx, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(got, expectedOrders,
		cmpopts.IgnoreFields(Order{}, "Date"))
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestRotateOrderHistory(t *testing.T) {
	ctx := context.Background()
	udb := &UserDB{DB: firestoreutil.StartTestEmulator(t, ctx),
		T: trace.NewNoopTracerProvider().Tracer("")}
	tu := testUser{id: "testuser", name: "abc"}
	if _, err := udb.EnsureAccountExists(ctx, tu); err != nil {
		t.Fatal(err)
	}

	// empty order history
	if err := udb.RotateOrderHistory(ctx, "testuser", 5); err != nil {
		t.Fatal(err)
	}

	// retain last orders
	ti := time.Date(2020, 04, 15, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 20; i++ {
		if err := udb.recordOrderHistory(ctx, "testuser",
			ti.Add(time.Second*time.Duration(i)), "", grpcoin.TradeAction_UNDEFINED, Amount{Units: int64(i)}, Amount{}); err != nil {
			t.Fatal(err)
		}
	}
	if err := udb.RotateOrderHistory(ctx, "testuser", 3); err != nil {
		t.Fatal(err)
	}
	hist, err := udb.UserOrderHistory(ctx, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	expected := []Order{
		{Date: ti.Add(17 * time.Second), Size: Amount{Units: 17}},
		{Date: ti.Add(18 * time.Second), Size: Amount{Units: 18}},
		{Date: ti.Add(19 * time.Second), Size: Amount{Units: 19}},
	}
	if diff := cmp.Diff(expected, hist); diff != "" {
		t.Fatal(diff)
	}
}

func TestAmount_IsNegative(t *testing.T) {
	tests := []struct {
		name   string
		fields Amount
		want   bool
	}{
		{name: "zero",
			want: false},
		{name: "zero part, neg nanos",
			fields: Amount{Nanos: -1},
			want:   true},
		{name: "zero part, pos nanos",
			fields: Amount{Nanos: 1},
			want:   false},
		{name: "zero nanos, neg units",
			fields: Amount{Units: -1},
			want:   true},
		{name: "zero nanos, pos units",
			fields: Amount{Units: 1},
			want:   false},
		{name: "pos both",
			fields: Amount{Units: 1, Nanos: 1},
			want:   false},
		{name: "neg both",
			fields: Amount{Units: -1, Nanos: -1},
			want:   true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.IsNegative(); got != tt.want {
				t.Errorf("Amount.IsNegative(%#v) = %v, want %v", tt.fields, got, tt.want)
			}
		})
	}
}

func TestAmount_IsZero(t *testing.T) {
	tests := []struct {
		name string
		in   Amount
		want bool
	}{
		{name: "zero",
			in:   Amount{0, 0},
			want: true},
		{name: "pos",
			in:   Amount{1, 1},
			want: false},
		{name: "neg",
			in:   Amount{-1, -1},
			want: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.IsZero(); got != tt.want {
				t.Errorf("Amount.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_batchDeleteAll(t *testing.T) {
	fs := firestoreutil.StartTestEmulator(t, context.TODO())
	col := fs.Collection("testcol")

	if err := batchDeleteAll(context.TODO(), fs, col.Documents(context.TODO())); err != nil {
		t.Fatal(err)
	}

	// add some documents
	for i := 0; i < 1002; i++ {
		_, _, err := col.Add(context.TODO(), map[string]string{"a": "b"})
		if err != nil {
			t.Fatal(err)
		}
	}
	it := col.Documents(context.TODO())
	if err := batchDeleteAll(context.TODO(), fs, it); err != nil {
		t.Fatal(err)
	}
	it = col.Documents(context.TODO())
	docs, err := it.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) > 0 {
		t.Fatalf("was not expecting results: got %d", len(docs))
	}
}
