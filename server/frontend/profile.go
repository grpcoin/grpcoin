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
	_ "embed"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpcoin/grpcoin/server/userdb"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (fe *Frontend) userProfile(w http.ResponseWriter, r *http.Request) error {
	uid := mux.Vars(r)["id"]
	if uid == "" {
		return status.Error(codes.InvalidArgument, "url does not have user id")
	}
	if s := trace.SpanFromContext(r.Context()); s != nil {
		s.SetAttributes(attribute.String("user.id", uid))
	}

	u, ok, err := fe.DB.Get(r.Context(), uid)
	if err != nil {
		return err
	} else if !ok {
		return status.Error(codes.NotFound, "user not found")
	}
	orders, err := fe.DB.UserOrderHistory(r.Context(), uid)
	if err != nil {
		return err
	}
	for i := 0; i < len(orders)/2; i++ {
		orders[i], orders[len(orders)-1-i] = orders[len(orders)-1-i], orders[i]
	}
	quotes, err := fe.getQuotes(r.Context())
	if err != nil {
		return err
	}

	hist, err := fe.DB.UserValuationHistory(r.Context(), u.ID)
	if err != nil {
		return err
	}
	pv := valuation(u.Portfolio, quotes)
	type returns struct {
		Label   string
		Percent userdb.Amount
	}
	returnPercentages := []returns{
		{"1 hour", findReturns(hist, pv, time.Hour)},
		{"6 hours", findReturns(hist, pv, time.Hour*6)},
		{"24 hours", findReturns(hist, pv, time.Hour*24)},
		{"1 week", findReturns(hist, pv, time.Hour*24*7)},
		{"30 days", findReturns(hist, pv, time.Hour*24*30)},
	}
	return tpl.Funcs(funcs).ExecuteTemplate(w, "profile.tpl", map[string]interface{}{
		"u":       u,
		"orders":  orders,
		"returns": returnPercentages,
		"quotes":  quotes})
}

func findReturns(history []userdb.ValuationHistory, currentValue userdb.Amount, ago time.Duration) userdb.Amount {
	// TODO decide whether to truncate here or not
	h := portfolioSnapshotAt(history, ago, time.Now())
	if h == nil {
		return userdb.Amount{}
	}
	return userdb.ToAmount(currentValue.F().Sub(h.Value.F()).Div(h.Value.F().Abs()).Mul(decimal.NewFromInt(100)))
}

func portfolioSnapshotAt(arr []userdb.ValuationHistory, ago time.Duration, now time.Time) *userdb.ValuationHistory {
	var last *userdb.ValuationHistory
	for j := len(arr) - 1; j >= 0; j-- {
		if now.Sub(arr[j].Date) > ago {
			break
		}
		last = &arr[j]
	}
	return last
}
