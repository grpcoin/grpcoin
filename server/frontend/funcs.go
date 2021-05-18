package frontend

import (
	"fmt"
	"text/template"

	"github.com/grpcoin/grpcoin/server/userdb"
)

var (
	funcs = template.FuncMap{
		"fmtAmount": fmtAmount,
		"fmtPrice":  fmtPrice,
	}
)

func fmtAmount(a userdb.Amount) string { return fmt.Sprintf("%d,%d", a.Units, a.Nanos) }
func fmtPrice(a userdb.Amount) string  { return fmt.Sprintf("%d,%02d", a.Units, a.Nanos/10000000) }
func valuation(p userdb.Portfolio, quotes map[string]userdb.Amount) userdb.Amount {
	total := p.CashUSD.F()
	for curr, amt := range p.Positions {
		// TODO we are not returning an error if quotes don't list the held currency
		total = total.Add(amt.F().Mul(quotes[curr].F()))
	}
	return userdb.ToAmount(total)
}
