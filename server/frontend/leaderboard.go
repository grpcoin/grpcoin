package frontend

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"text/template"

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
	tpl := `LEADERBOARD:
{{ range $i,$v := .users }}
{{ $i }}. {{$v.User.DisplayName}} (Valuation: USD {{rp $v.Valuation}}) (Cash: USD {{rp $v.User.Portfolio.CashUSD }})
{{- end }}`

	// TODO do not parse on every request
	t, err := template.New("").Funcs(funcs).Parse(tpl)
	if err != nil {
		return err
	}
	return t.Execute(w, map[string]interface{}{
		"users": resp})
}
