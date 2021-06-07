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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

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

	db, shutdownDB, err := serverutil.DetectDatabase(ctxzap.ToContext(ctx, log.With(zap.String("facility", "db"))),
		flTestData, onCloudRun, flRealData)
	defer shutdownDB()

	coinbaseQuotes := &coinbase.QuoteProvider{Logger: log}
	go coinbaseQuotes.Sync(ctx)

	trace, flushTraces := serverutil.GetTracer("grpcoin-frontend", onCloudRun)
	if err != nil {
		log.Fatal("failed to init tracing", zap.Error(err))
	}
	defer flushTraces(log.With(zap.String("facility", "tracing")))
	fe := frontend{
		QuoteProvider: coinbaseQuotes,
		QuoteDeadline: time.Second,
		CronSAEmail:   os.Getenv("CRON_SERVICE_ACCOUNT"),
		Trace:         trace,
		DB: &userdb.UserDB{
			DB: db,
			T:  trace}}

	addr := os.Getenv("LISTEN_ADDR")
	port := os.Getenv("PORT")
	server := &http.Server{
		Handler: fe.Handler(log),
		Addr:    addr + ":" + port,
	}
	log.Debug("starting to listen", zap.String("addr", addr+":"+port))
	go func() {
		<-ctx.Done()
		log.Debug("shutdown signal received")
		server.Shutdown(context.TODO())
	}()

	err = server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		log.Debug("http server closed")
	} else {
		log.Fatal("http server failed", zap.Error(err))
	}
	log.Debug("finished http server")
}
