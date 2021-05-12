package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/ahmetb/grpcoin/auth/github"
)

func TestAuthInfoFromContext(t *testing.T) {
	ctx := context.Background()

	if v := AuthInfoFromContext(ctx); v != nil {
		t.Fatal("expected nil")
	}

	ctx = context.WithValue(ctx, ctxAuthUserInfo{}, github.GitHubUser{})
	v := AuthInfoFromContext(ctx)
	if v == nil {
		t.Fatal("not expected nil")
	}
}

func TestAuthenticatingInterceptor(t *testing.T) {
	got, err := AuthenticatingInterceptor(MockAuthenticator{
		func(c context.Context) (AuthenticatedUser, error) { return &github.GitHubUser{}, nil },
	})(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if u := AuthInfoFromContext(got); u == nil {
		t.Fatal("auth info did not propagate into returned context")
	}

	_, err = AuthenticatingInterceptor(MockAuthenticator{
		func(c context.Context) (AuthenticatedUser, error) { return nil, fmt.Errorf("some error") },
	})(context.Background())
	if err == nil {
		t.Fatal("did not get the error from auth func")
	}
}
