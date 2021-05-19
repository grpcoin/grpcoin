package frontend

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/grpcoin/grpcoin/server/userdb"
	"github.com/hako/durafmt"
)

var (
	funcs = template.FuncMap{
		"fmtAmount":   fmtAmount,
		"fmtPrice":    fmtPrice,
		"fmtDate":     fmtDate,
		"fmtDuration": fmtDuration,
		"pv":          valuation,
		"since":       since,
	}
)

func fmtAmount(a userdb.Amount) string {
	return strings.TrimRight(fmt.Sprintf("%d,%09d", a.Units, a.Nanos), "0")
}

func fmtPrice(a userdb.Amount) string { return fmt.Sprintf("%d,%02d", a.Units, a.Nanos/10000000) }

func valuation(p userdb.Portfolio, quotes map[string]userdb.Amount) userdb.Amount {
	total := p.CashUSD.F()
	for curr, amt := range p.Positions {
		// TODO we are not returning an error if quotes don't list the held currency
		total = total.Add(amt.F().Mul(quotes[curr].F()))
	}
	return userdb.ToAmount(total)
}

func fmtDate(t time.Time) string { return t.Truncate(time.Hour * 24).Format("2 January 2006") }

func since(t time.Time) time.Duration { return time.Since(t) }

func fmtDuration(t time.Duration, maxUnits int) string {
	return durafmt.Parse(t).LimitFirstN(maxUnits).String()
}
