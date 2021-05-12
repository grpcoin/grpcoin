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
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/firestore"
	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb "github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/auth/github"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := os.Getenv("LISTEN_ADDR")

	lis, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		panic(err)
	}
	var rc *redis.Client
	if r := os.Getenv("REDIS_IP"); r == "" {
		rc = dummyRedis()
	} else {
		rc = redis.NewClient(&redis.Options{Addr: r + ":6379"})
	}
	if err := rc.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	fs, err := firestore.NewClient(ctx, "grpcoin")
	if err != nil {
		panic(err)
	}

	ac := &AccountCache{cache: rc}
	udb := &userdb.UserDB{DB: fs}
	as := &accountService{cache: ac, udb: udb}
	au := &github.GitHubAuthenticator{}
	ts := &tickerService{}
	err = prepServer(ctx, au, udb, as, ts).Serve(lis)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("gracefully shut down the server")
}

func prepServer(ctx context.Context, au auth.Authenticator, udb *userdb.UserDB, as *accountService, ts *tickerService) *grpc.Server {
	logOpts := []grpc_zap.Option{}
	zapLogger, err := zap.NewProduction() // make an arg
	if err != nil {
		panic(err)
	}

	unaryInterceptors := grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(zapLogger, logOpts...),
		grpc_auth.UnaryServerInterceptor(auth.AuthenticatingInterceptor(au)),
		grpc_auth.UnaryServerInterceptor(udb.EnsureAccountExistsInterceptor()),
	)
	streamInterceptors := grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(zapLogger, logOpts...),
	)
	// grpc_zap.ReplaceGrpcLoggerV2(zapLogger)
	srv := grpc.NewServer(unaryInterceptors, streamInterceptors)
	pb.RegisterAccountServer(srv, as)
	pb.RegisterTickerInfoServer(srv, ts) // this one is not authenticated (since it's stream-only, no unary)
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	return srv
}
