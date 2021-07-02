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
	_ "embed"
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpcoin/grpcoin/userdb"
)

func (fe *frontend) getQuotes(ctx context.Context) (map[string]userdb.Amount, error) {
	ctx, s := fe.Trace.Start(ctx, "realtime quote")
	defer s.End()

	var mu sync.Mutex
	out := make(map[string]userdb.Amount)

	eg, _ := errgroup.WithContext(ctx)
	for _, s := range fe.SupportedSymbols {
		quote := s // NB: needed to capture for the closure below
		eg.Go(func() error {
			v, err := fe.QuoteProvider.GetQuote(ctx, quote)
			if errors.Is(err, context.DeadlineExceeded) {
				return status.Errorf(codes.Unavailable, "could not get real-time market quote for %s in %v", quote, fe.QuoteDeadline)
			} else if err != nil {
				return status.Errorf(codes.Internal, "failed to retrieve a quote: %v", err)
			}
			mu.Lock()
			out[quote] = userdb.Amount{Units: v.GetUnits(), Nanos: v.GetNanos()}
			mu.Unlock()
			return nil
		})
	}
	return out, eg.Wait()
}

type leaderboardUser struct {
	User                userdb.User
	TotalPortfolioValue userdb.Amount
}

type LeaderboardHandlerData struct {
	Users            []leaderboardUser
	TotalTradeCount  int
	TotalTradeVolume float64
}

func (fe *frontend) leaderboard(w http.ResponseWriter, r *http.Request) error {
	quoteCtx, cancel := context.WithTimeout(r.Context(), fe.QuoteDeadline)
	defer cancel()
	quotes, err := fe.getQuotes(quoteCtx)
	if err != nil {
		return err
	}
	users, err := fe.DB.GetAll(r.Context())
	if err != nil {
		return err
	}

	var out LeaderboardHandlerData
	for _, u := range users {
		out.Users = append(out.Users, leaderboardUser{
			User:                u,
			TotalPortfolioValue: valuation(u.Portfolio, quotes)})
	}
	sort.Slice(out.Users, func(i, j int) bool {
		return !out.Users[i].TotalPortfolioValue.Less(out.Users[j].TotalPortfolioValue)
	})
	out.TotalTradeCount, err = fe.DB.TradeCounter.PastDayTradeCounts(r.Context(), time.Now())
	if err != nil {
		return err
	}
	out.TotalTradeVolume, err = fe.DB.TradeCounter.PastDayTradeVolume(r.Context(), time.Now())
	if err != nil {
		return err
	}
	return tpl.ExecuteTemplate(w, "leaderboard.tmpl", out)
}
