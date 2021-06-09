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
	"errors"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/apiserver/auth"
	"github.com/grpcoin/grpcoin/apiserver/auth/github"
	"github.com/grpcoin/grpcoin/apiserver/ratelimiter"
	"github.com/grpcoin/grpcoin/realtimequote/coinbase"
	"github.com/grpcoin/grpcoin/serverutil"
	"github.com/grpcoin/grpcoin/userdb"
)

var (
	flRealData bool
	flTestData string
)

func init() {
	flag.BoolVar(&flRealData, "use-real-db", false, "run against production database (requires $GOOGLE_CLOUD_PROJECT set), ignored when running on prod")
	flag.StringVar(&flTestData, "test-data", "testdata/local.db", "test data to load into the emulator when running locally, ignored when real db is used")
}

func main() {
	flag.Parse()
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	onCloudRun := os.Getenv("K_SERVICE") != ""

	log, err := serverutil.GetLogging(onCloudRun)
	if err != nil {
		panic(err)
	}

	tp, traceFlush := serverutil.GetTracer("grpcoin-api", onCloudRun)
	defer traceFlush(log.With(zap.String("facility", "trace")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	rc, close, err := serverutil.ConnectRedis(ctx, os.Getenv("REDIS_IP"))
	if err != nil {
		log.Fatal("failed to get redis instance", zap.Error(err))
	}
	defer close()
	defer rc.Close()

	db, shutdown, err := serverutil.DetectDatabase(ctxzap.ToContext(ctx, log.With(zap.String("facility", "db"))),
		flTestData, onCloudRun, flRealData)
	defer shutdown()

	ac := &AccountCache{cache: rc}
	udb := &userdb.UserDB{DB: db, T: tp}
	as := &accountService{cache: ac, udb: udb}
	au := &github.GitHubAuthenticator{T: tp, Cache: rc}
	cb := &coinbase.QuoteProvider{
		Logger: log.With(zap.String("facility", "coinbase"))}
	go cb.Sync(ctx)
	ts := &tickerService{}
	pt := &tradingService{udb: udb, tp: cb, tr: tp}
	rl := ratelimiter.New(rc, time.Now, tp)
	grpcServer := prepServer(log, au, rl, udb, as, ts, pt)
	listenHost := os.Getenv("LISTEN_ADDR")
	addr := net.JoinHostPort(listenHost, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("tcp listen failed", zap.Error(err))
	}

	log.Debug("starting to listen", zap.String("addr", addr))
	go func() {
		<-ctx.Done()
		log.Debug("shutdown signal received")
		grpcServer.GracefulStop()
	}()
	if err := grpcServer.Serve(lis); errors.Is(err, grpc.ErrServerStopped) {
		log.Debug("grpc: shut down the server")
	} else {
		log.Fatal("grpc: server failed", zap.Error(err))
	}
}

func prepServer(log *zap.Logger, au auth.Authenticator, rl ratelimiter.RateLimiter, udb *userdb.UserDB, as *accountService, ts *tickerService, pt *tradingService) *grpc.Server {
	unaryInterceptors := grpc_middleware.WithUnaryServerChain(
		otelgrpc.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(log),
		grpc_auth.UnaryServerInterceptor(auth.AuthenticatingInterceptor(au)),
		grpc_auth.UnaryServerInterceptor(rateLimitInterceptor(rl)),
		grpc_auth.UnaryServerInterceptor(udb.EnsureAccountExistsInterceptor()),
	)

	// not adding the otel interceptor here since it's just the TickerInfo.Watch() call for now
	streamInterceptors := grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(log),
		grpc_auth.StreamServerInterceptor(rateLimitInterceptor(rl)),
	)
	// grpc_zap.ReplaceGrpcLoggerV2(log)
	srv := grpc.NewServer(unaryInterceptors, streamInterceptors)
	pb.RegisterAccountServer(srv, as)
	pb.RegisterTickerInfoServer(srv, ts) // this one is not authenticated (since it's stream-only, no unary)
	pb.RegisterPaperTradeServer(srv, pt)
	return srv
}
