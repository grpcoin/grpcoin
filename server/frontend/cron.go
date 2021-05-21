package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/grpcoin/grpcoin/server/userdb"
	"go.uber.org/zap"
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

	// TODO process in parallel in batches
	for _, u := range users {
		pv := valuation(u.Portfolio, quotes)
		if err := fe.DB.SetUserValuationHistory(r.Context(), u.ID, userdb.ValuationHistory{
			Date:  t,
			Value: pv,
		}); err != nil {
			// TODO warning
			log.Warn("failed to process user", zap.String("id", u.ID), zap.Error(err))
			continue
		}
		if err := fe.DB.RotateUserValuationHistory(r.Context(), u.ID,
			t.Add(-maxPortfolioHistory)); err != nil {
			log.Warn("rotation history failed", zap.String("id", u.ID), zap.Error(err))
			continue
		}
		log.Debug("processed user", zap.String("id", u.ID))
	}
	return nil
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
