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
	"bufio"
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"html/template"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpcoin/grpcoin/realtimequote"
	"github.com/grpcoin/grpcoin/realtimequote/fanout"
	"github.com/grpcoin/grpcoin/userdb"
	"github.com/purini-to/zapmw"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	//go:embed templates/*
	templateFS embed.FS
	tpl        = template.Must(template.New("").Funcs(funcs).ParseFS(templateFS,
		"templates/*.tmpl"))
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type frontend struct {
	QuoteProvider realtimequote.QuoteProvider
	QuoteDeadline time.Duration
	QuoteFanout   *fanout.QuoteFanoutService

	CronSAEmail string // email for the SA allowed to run cron endpoints

	Trace trace.Tracer
	DB    *userdb.UserDB
	Redis *redis.Client
}

func (fe *frontend) Handler(log *zap.Logger) http.Handler {
	m := mux.NewRouter()
	m.Use(handlers.ProxyHeaders,
		handlers.CompressHandler,
		otelmux.Middleware("grpcoin-frontend"),
		zapmw.WithZap(log, withStackdriverFields),
		zapmw.Request(zapcore.InfoLevel, "request"),
		zapmw.Recoverer(zapcore.ErrorLevel, "recover", zapmw.RecovererDefault))
	m.HandleFunc("/_cron/pv", toHandler(fe.calcPortfolioHistory))
	m.HandleFunc("/api/portfolioValuation/{id}", toHandler(fe.apiPortfolioHistory))
	m.HandleFunc("/user/{id}", toHandler(fe.userProfile))
	m.HandleFunc("/ws/tickers", toHandler(fe.wsTickers))
	m.HandleFunc("/", toHandler(fe.leaderboard))
	return m
}

// toHandler allows handlers to return errors, ideally as grpc/status errors
// which are converted to HTTP statuses and properly rendered error responses.
func toHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bw := &bufferedRespWriter{ResponseWriter: w}
		err := f(bw, r)
		if err == nil {
			if bw.status != 0 {
				w.WriteHeader(bw.status)
			}
			io.Copy(w, &bw.b)
			return
		}
		err = status.FromContextError(err).Err() // convert ctx errors to grpc status when possible
		handleErr(loggerFrom(r.Context()), w, bw.status, err)
	}
}

type respErr struct {
	Status  int    `json:"status,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	ID      string `json:"id,omitempty"`
}

func handleErr(log *zap.Logger, w http.ResponseWriter, statusOverride int, err error) {
	var outErr respErr
	id := uuid.New().String()
	grpcStatus, ok := status.FromError(err)
	log.Error("request error", zap.Error(err), zap.String("error.id", id))
	if !ok {
		outErr = respErr{
			ID:      id,
			Status:  http.StatusInternalServerError,
			Code:    codes.Internal.String(),
			Message: err.Error()}
		if statusOverride != 0 {
			outErr.Status = statusOverride
		}
	}
	outErr = respErr{
		ID:      id,
		Status:  runtime.HTTPStatusFromCode(grpcStatus.Code()),
		Code:    grpcStatus.Code().String(),
		Message: grpcStatus.Message()}
	w.WriteHeader(outErr.Status)
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	e.Encode(outErr)
}

type bufferedRespWriter struct {
	b      bytes.Buffer
	status int
	http.ResponseWriter
}

func (bw *bufferedRespWriter) Write(d []byte) (int, error) {
	return bw.b.Write(d)
}

func (bw *bufferedRespWriter) WriteHeader(statusCode int) {
	bw.status = statusCode
}

func (bw *bufferedRespWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return bw.ResponseWriter.(http.Hijacker).Hijack()
}
