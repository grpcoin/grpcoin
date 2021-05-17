package main

import (
	"fmt"
	"net/http"

	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type webHandler struct {
	qp  *coinbaseQuoteProvider
	tp  trace.Tracer
	udb *userdb.UserDB
}

func (_ *webHandler) health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func (w *webHandler) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", w.health)
	return otelhttp.NewHandler(mux, "frontend")
}
