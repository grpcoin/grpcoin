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

	firestore "cloud.google.com/go/firestore"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/firestoretestutil"
)

type testUser struct {
	id   string
	name string
}

func (t testUser) DBKey() string       { return t.id }
func (t testUser) DisplayName() string { return t.name }
func (t testUser) ProfileURL() string  { return "https://" + t.name }

func TestGetUser_notFound(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx)}
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
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx)}
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
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	udb := &UserDB{DB: firestoretestutil.StartEmulator(t, ctx)}
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

func TestUserDB_Trade(t *testing.T) {
	type fields struct {
		DB *firestore.Client
	}
	type args struct {
		ctx      context.Context
		uid      string
		ticker   string
		action   grpcoin.TradeAction
		quote    *grpcoin.Amount
		quantity *grpcoin.Amount
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserDB{
				DB: tt.fields.DB,
			}
			if err := u.Trade(tt.args.ctx, tt.args.uid, tt.args.ticker, tt.args.action, tt.args.quote, tt.args.quantity); (err != nil) != tt.wantErr {
				t.Errorf("UserDB.Trade() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
