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

package frontend

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpcoin/grpcoin/server/realtimequote"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	//go:embed templates/*
	templateFS embed.FS
	tpl        = template.Must(template.New("").Funcs(funcs).ParseFS(templateFS,
		"templates/*.tpl"))
)

type Frontend struct {
	QuoteProvider realtimequote.QuoteProvider
	QuoteDeadline time.Duration

	CronSAEmail string // email for the SA allowed to run cron endpoints

	Trace trace.Tracer
	DB    *userdb.UserDB
}

func (_ *Frontend) health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}
func (fe *Frontend) Handler(log *zap.Logger) http.Handler {
	m := mux.NewRouter()
	m.Use(otelmux.Middleware("grpcoin-frontend"))
	m.Use(withLogging(log))
	m.HandleFunc("/health", fe.health)
	m.HandleFunc("/_cron/pv", toHandler(fe.calcPortfolioHistory))
	m.HandleFunc("/user/{id}", toHandler(fe.userProfile))
	m.HandleFunc("/", toHandler(fe.leaderboard))
	return m
}

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
		if errors.Is(err, context.Canceled) {
			// convert context cancellations into proper grpc Canceled error
			err = status.Error(codes.Canceled, err.Error())
		}
		grpcStatus, ok := status.FromError(err)
		if !ok {
			if bw.status != 0 {
				w.WriteHeader(bw.status)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		w.WriteHeader(runtime.HTTPStatusFromCode(grpcStatus.Code()))
		fmt.Fprintf(w, `ERROR: code: %s -- message: %s`, grpcStatus.Code(), grpcStatus.Message())
	}
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
