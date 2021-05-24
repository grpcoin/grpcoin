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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/grpcoin/grpcoin/server/userdb"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	maxPortfolioHistory = time.Hour * 24 * 31
)

func (fe *Frontend) calcPortfolioHistory(w http.ResponseWriter, r *http.Request) error {
	log := loggerFrom(r.Context())
	if fe.CronSAEmail != "" {
		token := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
		if err := verifyJWT(r.Context(), fe.CronSAEmail, token); err != nil {
			return status.Error(codes.PermissionDenied, err.Error())
		}
		log.Info("jwt verification passed")
	}

	subCtx, s := fe.Trace.Start(r.Context(), "retrieve users")
	users, err := fe.DB.GetAll(subCtx)
	if err != nil {
		return err
	}
	s.End()

	t := time.Now().UTC().Truncate(time.Hour)
	quotes, err := fe.getQuotes(r.Context())
	if err != nil {
		return err
	}

	var batchSize int64 = 10
	sem := semaphore.NewWeighted(batchSize)

	// TODO process in parallel in batches
	for _, u := range users {
		sem.Acquire(context.TODO(), 1)
		go func(u userdb.User) {
			defer sem.Release(1)
			pv := valuation(u.Portfolio, quotes)
			if err := fe.DB.SetUserValuationHistory(r.Context(), u.ID, userdb.ValuationHistory{
				Date:  t,
				Value: pv,
			}); err != nil {
				log.Warn("failed to process user", zap.String("id", u.ID), zap.Error(err))
				return
			}
			if err := fe.DB.RotateUserValuationHistory(r.Context(), u.ID,
				t.Add(-maxPortfolioHistory)); err != nil {
				log.Warn("rotation history failed", zap.String("id", u.ID), zap.Error(err))
				return
			}
			log.Debug("processed user", zap.String("id", u.ID))
		}(u)
	}
	return sem.Acquire(r.Context(), batchSize)
}

func verifyJWT(ctx context.Context, expectedSAEmail, token string) error {
	p, err := idtoken.Validate(ctx, token, "")
	if err != nil {
		return err
	}
	em := p.Claims["email"]
	if em != expectedSAEmail {
		return fmt.Errorf("request identity %q is not the expected account", em)
	}
	return nil
}
