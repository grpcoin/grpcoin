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

	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	ratelimiter2 "github.com/grpcoin/grpcoin/ratelimiter"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/binance"
	"github.com/grpcoin/grpcoin/realtimequote/fanout"
	"github.com/grpcoin/grpcoin/tradecounters"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/apiserver/auth"
	"github.com/grpcoin/grpcoin/apiserver/auth/github"
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
	rc, closeInstance, err := serverutil.ConnectRedis(ctx, os.Getenv("REDIS_IP"))
	if err != nil {
		log.Fatal("failed to get redis instance", zap.Error(err))
	}
	defer closeInstance()
	defer rc.Close()

	db, shutdown, err := serverutil.DetectDatabase(ctxzap.ToContext(ctx, log.With(zap.String("facility", "db"))),
		flTestData, onCloudRun, flRealData)
	defer shutdown()

	accountCache := &AccountCache{cache: rc}
	udb := &userdb.UserDB{
		DB:           db,
		T:            tp,
		Cache:        userdb.UserDBCache{R: rc},
		TradeCounter: &tradecounters.TradeCounter{DB: rc}}
	accountSvc := &accountService{cache: accountCache, udb: udb}
	authenticator := &github.GitHubAuthenticator{T: tp, Cache: rc}

	quoteStream := realtimequote.QuoteStreamFunc(binance.WatchSymbols)
	supportedTickers := realtimequote.SupportedTickers
	quoteProvider := realtimequote.NewReconnectingQuoteProvider(ctx,
		log.With(zap.String("facility", "quotes")),
		quoteStream,
		supportedTickers...)
	tickerSvc := &tickerService{
		maxRate:          time.Millisecond * 100,
		supportedTickers: supportedTickers,
		fanout: fanout.NewQuoteFanoutService(func(ctx context.Context) (<-chan realtimequote.Quote, error) {
			return quoteStream(ctx, supportedTickers...)
		})}
	tradingSvc := &tradingService{
		udb:              udb,
		quoteProvider:    quoteProvider,
		supportedTickers: supportedTickers,
		tracer:           tp}
	rl := ratelimiter2.New(rc, time.Now, tp, time.Minute)
	grpcServer := prepServer(log, authenticator, rl, udb, accountSvc, tickerSvc, tradingSvc)
	host := os.Getenv("LISTEN_ADDR")
	addr := net.JoinHostPort(host, port)
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
	if err := grpcServer.Serve(lis); err == nil || errors.Is(err, grpc.ErrServerStopped) {
		log.Debug("grpc: shut down the server")
	} else {
		log.Fatal("grpc: server failed", zap.Error(err))
	}
}

func prepServer(log *zap.Logger, au auth.Authenticator, rl ratelimiter2.RateLimiter, udb *userdb.UserDB, as *accountService, ts *tickerService, pt *tradingService) *grpc.Server {
	unaryInterceptors := grpcmiddleware.WithUnaryServerChain(
		otelgrpc.UnaryServerInterceptor(),
		grpcctxtags.UnaryServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
		grpczap.UnaryServerInterceptor(log),
		internalErrorHidingInterceptor,
		grpcauth.UnaryServerInterceptor(auth.AuthenticatingInterceptor(au)),
		grpcauth.UnaryServerInterceptor(rateLimitInterceptor(rl)),
		grpcauth.UnaryServerInterceptor(udb.EnsureAccountExistsInterceptor()),
	)

	// not adding the otel interceptor here since it's just the TickerInfo.Watch() call for now
	streamInterceptors := grpcmiddleware.WithStreamServerChain(
		grpcctxtags.StreamServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
		grpczap.StreamServerInterceptor(log),
		grpcauth.StreamServerInterceptor(rateLimitInterceptor(rl)),
	)
	//grpc_zap.ReplaceGrpcLoggerV2(log) // grpc's internal logs
	srv := grpc.NewServer(unaryInterceptors, streamInterceptors)
	pb.RegisterAccountServer(srv, as)
	pb.RegisterTickerInfoServer(srv, ts) // this one is not authenticated (since it's stream-only, no unary)
	pb.RegisterPaperTradeServer(srv, pt)
	return srv
}

func internalErrorHidingInterceptor(ctx context.Context,
	req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	c := status.Code(err)
	if c == codes.Internal {
		id := uuid.New().String()
		newErr := status.Errorf(c, "internal error occurred (for debugging purposes, error.id=%s)", id)
		status.Convert(err).Details()
		ctxzap.Extract(ctx).Error("internal error", zap.Error(err), zap.String("error.id", id))
		return resp, newErr
	}
	return resp, err
}
