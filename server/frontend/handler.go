package frontend

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpcoin/grpcoin/server/realtimequote"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/status"
)

type Frontend struct {
	QuoteProvider realtimequote.QuoteProvider
	QuoteDeadline time.Duration

	Trace trace.Tracer
	DB    *userdb.UserDB
}

func (_ *Frontend) health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}
func (fe *Frontend) Handler() http.Handler {
	m := mux.NewRouter()
	m.Use(otelmux.Middleware("grpcoin-frontend"))
	m.HandleFunc("/health", fe.health)
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
