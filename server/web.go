package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpcoin/grpcoin/server/userdb"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type webHandler struct {
	qp  *coinbaseQuoteProvider
	tp  trace.Tracer
	udb *userdb.UserDB
}

func (_ *webHandler) health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

type leaderboardUser struct {
	U         userdb.User
	Valuation userdb.Amount
}

type leaderboardResp []leaderboardUser

func (l leaderboardResp) Len() int          { return len(l) }
func (l leaderboardResp) Swap(i int, j int) { l[i], l[j] = l[j], l[i] }
func (l leaderboardResp) Less(i int, j int) bool {
	if l[i].Valuation.Units < l[j].Valuation.Units {
		return true
	} else if l[i].Valuation.Units == l[j].Valuation.Units && l[i].Valuation.Nanos < l[j].Valuation.Nanos {
		return true
	}
	return false
}

func (wh *webHandler) leaderboard(w http.ResponseWriter, r *http.Request) error {
	// get real-time BTC quote
	subCtx, s := wh.tp.Start(r.Context(), "get quote")
	quoteCtx, cancel := context.WithTimeout(subCtx, quoteDeadline)
	defer cancel()
	ticker := "BTC"
	btcQuote, err := wh.qp.GetQuote(quoteCtx, ticker)
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Errorf(codes.Unavailable, "could not get real-time market quote for %s in %v", ticker, quoteDeadline)
	} else if err != nil {
		return status.Errorf(codes.Internal, "failed to retrieve a quote: %v", err)
	}
	s.End()

	// get all users (+portfolio)
	quotes := map[string]userdb.Amount{
		"BTC": {Units: btcQuote.GetUnits(), Nanos: btcQuote.GetNanos()},
	}
	users, err := wh.udb.GetAll(r.Context())
	if err != nil {
		return err
	}
	var resp leaderboardResp
	for _, u := range users {
		resp = append(resp, leaderboardUser{U: u, Valuation: u.Portfolio.Valuation(quotes)})
	}
	sort.Sort(sort.Reverse(resp))
	tpl := `LEADERBOARD:
{{ range $i,$v := .users }}
{{ $i }}. {{$v.U.DisplayName}} (Valuation: USD {{rp $v.Valuation}}) (Cash: USD {{rp $v.U.Portfolio.CashUSD }})
{{- end }}`
	// TODO do not parse on every request
	t, err := template.New("").Funcs(template.FuncMap{
		"rp": func(a userdb.Amount) string {
			return fmt.Sprintf("%d,%02d", a.Units, a.Nanos/10000000)
		}}).Parse(tpl)
	if err != nil {
		return err
	}
	return t.Execute(w, map[string]interface{}{
		"users": resp})
	// calculate valuation
	// sort
	// optional: 24h change, 7d change

	/*
		   terraform --> cloud scheduler (cron trigger)
		   --> /_cron/recalculate-portfolio-value
		       --> foreach user:
		   	      {
		   			  id: asdfasdf
		   			  portfolio:{
		   				cash: xx.yy
		   				positions: [{btc, x.y}, {eth, z.j}]
		   			  }
		   			  daily: {
		   				[t1, x],
		   				[t2, x],
		   				[t3, x],
		   			  } 30d
					  hourly: 96h
		   		  }

	*/
}

func (w *webHandler) handler() http.Handler {
	m := mux.NewRouter()
	m.Use(otelmux.Middleware("grpcoin-frontend"))
	m.HandleFunc("/health", w.health)
	m.HandleFunc("/", toHandler(w.leaderboard))
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
