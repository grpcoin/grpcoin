package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testUser struct {
	id   string
	name string
}

func (t testUser) DBKey() string       { return t.id }
func (t testUser) DisplayName() string { return t.name }
func (t testUser) ProfileURL() string  { return "https://" + t.name }

var _ AuthenticatedUser = testUser{}

func TestGetUser_notFound(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	fs := startFirebaseEmulator(t, ctx)
	udb := &userDB{fs: fs}
	tu := testUser{id: "foo"}

	u, ok, err := udb.get(ctx, tu)
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
	fs := startFirebaseEmulator(t, ctx)
	udb := &userDB{fs: fs}
	tu := testUser{id: "foobar", name: "ab"}

	err := udb.create(ctx, tu)
	if err != nil {
		t.Fatal(err)
	}

	uv, ok, err := udb.get(ctx, tu)
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
	}
	if diff := cmp.Diff(uv, expected); diff != "" {
		t.Fatal(diff)
	}
}
