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
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb "github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/grpcoin/grpcoin/server/auth"
	"github.com/grpcoin/grpcoin/server/auth/github"
	"github.com/grpcoin/grpcoin/server/userdb"
	stackdriver "github.com/tommy351/zap-stackdriver"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	onCloudRun := os.Getenv("K_SERVICE") != ""

	var log *zap.Logger
	if onCloudRun {
		c := zap.NewProductionConfig()
		c.EncoderConfig = stackdriver.EncoderConfig
		c.OutputPaths = []string{"stdout"}
		z, err := c.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return &stackdriver.Core{
				Core: core,
			}
		}), zap.Fields(
			stackdriver.LogServiceContext(&stackdriver.ServiceContext{
				Service: os.Getenv("K_SERVICE"),
				Version: os.Getenv("K_REVISION"),
			}),
		))
		if err != nil {
			panic(err)
		}
		log = z
	} else {
		z, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		z = z.With(zap.String("env", "dev"))
		log = z
	}

	var traceExporter trace.SpanExporter
	if onCloudRun {
		gcp, err := texporter.NewExporter()
		if err != nil {
			log.Fatal("failed to initialize gcp trace exporter", zap.Error(err))
		}
		traceExporter = gcp
	} else {
		traceExporter = dummyTraceExporter{}
	}
	tracer := trace.NewTracerProvider(trace.WithBatcher(traceExporter,
		trace.WithBatchTimeout(time.Millisecond*100),
		trace.WithMaxExportBatchSize(10)),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tracer)
	tp := otel.GetTracerProvider().Tracer("main")
	defer func() {
		log.Debug("force flushing trace spans")
		// we don't use the main ctx here as it'll be cancelled by the time this is executed
		if err := tracer.ForceFlush(context.TODO()); err != nil {
			log.Warn("failed to flush tracer", zap.Error(err))
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	var rc *redis.Client
	if r := os.Getenv("REDIS_IP"); r == "" {
		rc = dummyRedis()
	} else {
		rc = redis.NewClient(&redis.Options{Addr: r + ":6379"})
	}
	if err := rc.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping failed", zap.Error(err))
	}

	var proj string
	if onCloudRun {
		proj = firestore.DetectProjectID
	} else {
		proj = "grpcoin" // TODO do not hardcode this for local testing, maybe start fs emulator
	}
	fs, err := firestore.NewClient(ctx, proj)
	if err != nil {
		log.Fatal("failed to initialize firestore client", zap.String("project", proj), zap.Error(err))
	}

	ac := &AccountCache{cache: rc}
	udb := &userdb.UserDB{DB: fs, T: tp}
	as := &accountService{cache: ac, udb: udb}
	au := &github.GitHubAuthenticator{T: tp, Cache: rc}
	cb := &coinbaseQuoteProvider{}
	go cb.sync(ctx, "BTC")
	ts := &tickerService{}
	pt := &tradingService{udb: udb, tp: cb, tr: tp}
	grpcServer := prepServer(ctx, log, au, udb, as, ts, pt)

	// implement mux here
	frontendHandler := (&webHandler{tp: tp, udb: udb, qp: cb}).handler()
	mux := newHTTPandGRPCMux(frontendHandler, grpcServer)
	http2Server := &http2.Server{}
	http1Server := &http.Server{Handler: h2c.NewHandler(mux, http2Server)}

	// listener --> h2c converter
	//				--> grpc: grpc.ServeHTTP
	//				--> h1

	addr := os.Getenv("LISTEN_ADDR")
	lis, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		log.Fatal("tcp listen failed", zap.Error(err))
	}

	log.Debug("starting to listen", zap.String("addr", addr+":"+port))
	go func() {
		<-ctx.Done()
		log.Debug("shutdown signal received")
		// TODO using grpcServer.ServeHTTP above does not make use of grpc-go's
		// http2 server, and here Shutdown() does not keep track of hijacked
		// connections. so this is not really a graceful termination.
		// Find a way to fix that.
		http1Server.Shutdown(ctx)
	}()
	if err := http1Server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("http: server failed", zap.Error(err))
	}
	log.Debug("http: shut down the server")
}

func prepServer(ctx context.Context, log *zap.Logger, au auth.Authenticator,
	udb *userdb.UserDB, as *accountService, ts *tickerService,
	pt *tradingService) *grpc.Server {
	unaryInterceptors := grpc_middleware.WithUnaryServerChain(
		otelgrpc.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(log),
		grpc_auth.UnaryServerInterceptor(auth.AuthenticatingInterceptor(au)),
		grpc_auth.UnaryServerInterceptor(udb.EnsureAccountExistsInterceptor()),
	)

	// not adding the otel interceptor here since it's just the TickerInfo.Watch() call for now
	streamInterceptors := grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(log),
	)
	// grpc_zap.ReplaceGrpcLoggerV2(log)
	srv := grpc.NewServer(unaryInterceptors, streamInterceptors)
	pb.RegisterAccountServer(srv, as)
	pb.RegisterTickerInfoServer(srv, ts) // this one is not authenticated (since it's stream-only, no unary)
	pb.RegisterPaperTradeServer(srv, pt)
	return srv
}

func newHTTPandGRPCMux(httpHand http.Handler, grpcHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("content-type"), "application/grpc") {
			grpcHandler.ServeHTTP(w, r)
			return
		}
		httpHand.ServeHTTP(w, r)
	})
}
