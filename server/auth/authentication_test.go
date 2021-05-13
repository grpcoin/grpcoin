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
	"fmt"
	"testing"

	"github.com/grpcoin/grpcoin/auth/github"
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
