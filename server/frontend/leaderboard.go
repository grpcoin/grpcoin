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
	"context"
	_ "embed"
	"errors"
	"net/http"
	"sort"

	"github.com/grpcoin/grpcoin/server/userdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type leaderboardUser struct {
	User      userdb.User
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

func (fe *Frontend) getQuotes(ctx context.Context) (map[string]userdb.Amount, error) {
	subCtx, s := fe.Trace.Start(ctx, "realtime quote")
	defer s.End()
	quoteCtx, cancel := context.WithTimeout(subCtx, fe.QuoteDeadline)
	defer cancel()

	btcQuote, err := fe.QuoteProvider.GetQuote(quoteCtx, "BTC")
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, status.Errorf(codes.Unavailable, "could not get real-time market quote for %s in %v", "BTC", fe.QuoteDeadline)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve a quote: %v", err)
	}

	return map[string]userdb.Amount{
		"BTC": {Units: btcQuote.GetUnits(), Nanos: btcQuote.GetNanos()}}, nil
}

func (fe *Frontend) leaderboard(w http.ResponseWriter, r *http.Request) error {
	quotes, err := fe.getQuotes(r.Context())
	if err != nil {
		return err
	}
	users, err := fe.DB.GetAll(r.Context())
	if err != nil {
		return err
	}
	var resp leaderboardResp
	for _, u := range users {
		resp = append(resp, leaderboardUser{
			User:      u,
			Valuation: valuation(u.Portfolio, quotes)})
	}
	sort.Sort(sort.Reverse(resp))

	// TODO do not parse on every request
	return tpl.ExecuteTemplate(w, "leaderboard.tpl", map[string]interface{}{
		"users": resp})
}
