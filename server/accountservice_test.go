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

package main

import (
	"context"
	"net"
	"testing"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/auth/github"
	"github.com/grpcoin/grpcoin/server/firestoretestutil"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestTestAuth(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	fs := firestoretestutil.StartEmulator(t, context.TODO())
	au := auth.MockAuthenticator{
		F: func(c context.Context) (auth.AuthenticatedUser, error) {
			return &github.GitHubUser{ID: 1, Username: "abc"}, nil
		},
	}
	udb := &userdb.UserDB{DB: fs, T: trace.NewNoopTracerProvider().Tracer("")}
	lg, _ := zap.NewDevelopment()
	srv := prepServer(context.TODO(), lg, au, udb, &accountService{cache: &AccountCache{cache: dummyRedis()}}, nil, nil)
	go srv.Serve(l)
	defer srv.Stop()
	defer l.Close()

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client := grpcoin.NewAccountClient(cc)

	resp, err := client.TestAuth(context.TODO(), &grpcoin.TestAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetUserId() == "" {
		t.Fatal("got empty user id in auth response")
	}
	expected := "github_1"
	if resp.GetUserId() != expected {
		t.Fatalf("uid expected=%q got=%q", expected, resp.GetUserId())
	}
}
