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

package github

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGitHubAuthenticator(t *testing.T) {
	gh := &GitHubAuthenticator{}
	ctx := metadata.NewIncomingContext(context.Background(), nil)
	_, err := gh.Authenticate(ctx)
	if err == nil {
		t.Fatal("expected err without any creds")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatal("not a grpc status!")
	}
	if s.Code() != codes.Unauthenticated {
		t.Fatalf("got code: %v; expected Unauthenticated -- err:%v", s.Code(), err)
	}

	md := metadata.New(map[string]string{"authorization": "bad format"})
	ctx = metadata.NewIncomingContext(context.Background(), md)
	_, err = gh.Authenticate(ctx)
	if err == nil {
		t.Fatal("expected err with bad format")
	}
	s, ok = status.FromError(err)
	if !ok {
		t.Fatal("not a grpc status!")
	}
	if s.Code() != codes.InvalidArgument {
		t.Fatalf("got code: %v; expected InvalidArgument", s.Code())
	}

	md = metadata.New(map[string]string{"authorization": "Bearer 123"})
	ctx = metadata.NewIncomingContext(context.Background(), md)
	_, err = gh.Authenticate(ctx)
	if err == nil {
		t.Fatal("expected err with bad creds")
	}
	s, ok = status.FromError(err)
	if !ok {
		t.Fatal("not a grpc status!")
	}
	if s.Code() != codes.PermissionDenied {
		t.Fatalf("got code: %v; expected PermissionDenied: %v", s.Code(), err)
	}
}
