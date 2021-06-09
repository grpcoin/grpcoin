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
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"

	"github.com/grpcoin/grpcoin/userdb"
)

var (
	funcs = template.FuncMap{
		"fmtAmount":    fmtAmount,
		"fmtAmountRaw": fmtAmountRaw,
		"fmtPrice":     fmtPrice,
		"fmtPriceFull": fmtPriceFull,
		"fmtDate":      fmtDate,
		"fmtDuration":  fmtDuration,
		"fmtPercent":   fmtPercent,
		"toPercent":    toPercent,
		"pv":           valuation,
		"isNegative":   isNegative,
		"isZero":       userdb.Amount.IsZero,
		"since":        since,
		"mul":          mul,
		"div":          div,
		"profilePic":   profilePic,
	}
)

func fmtAmount(a userdb.Amount) string {
	v := fmt.Sprintf("%s.%09d", humanize.Comma(a.Units), a.Nanos)
	return trimTrailingZeros(v)
}

func fmtAmountRaw(a userdb.Amount) float64 {
	v, _ := a.F().Float64()
	return v
}

func trimTrailingZeros(v string) string {
	v = strings.TrimRight(v, "0")
	if strings.HasSuffix(v, ".") {
		v += "00"
	}
	return v
}

func fmtPrice(a userdb.Amount) string {
	return fmt.Sprintf("$%s.%02d", humanize.Comma(a.Units), a.Nanos/1_000_000_0)
}

func fmtPriceFull(a userdb.Amount) string {
	return trimTrailingZeros(fmt.Sprintf("$%s.%09d", humanize.Comma(a.Units), a.Nanos))
}

func fmtPercent(a userdb.Amount) string {
	nanos := a.Nanos
	if a.IsNegative() {
		nanos = -nanos
	}
	v := fmt.Sprintf("%d.%02d%%", a.Units, nanos/1_000_000_0)
	if a.IsNegative() && v[0] != '-' {
		v = "-" + v
	}
	return v
}

func valuation(p userdb.Portfolio, quotes map[string]userdb.Amount) userdb.Amount {
	total := p.CashUSD.F()
	for curr, amt := range p.Positions {
		// TODO we are not returning an error if quotes don't list the held currency
		total = total.Add(amt.F().Mul(quotes[curr].F()))
	}
	return userdb.ToAmount(total)
}

func mul(a, b userdb.Amount) userdb.Amount    { return userdb.ToAmount(a.F().Mul(b.F())) }
func div(a, b userdb.Amount) userdb.Amount    { return userdb.ToAmount(a.F().Div(b.F())) }
func toPercent(a userdb.Amount) userdb.Amount { return mul(a, userdb.Amount{Units: 100}) }
func isNegative(a userdb.Amount) bool         { return a.IsNegative() }

func fmtDate(t time.Time) string { return t.Truncate(time.Hour * 24).Format("2 January 2006") }

func since(t time.Time) time.Duration { return time.Since(t) }

func fmtDuration(t time.Duration, maxUnits int) string {
	return durafmt.Parse(t).LimitFirstN(maxUnits).String()
}

func profilePic(profileURL string) string {
	if strings.HasPrefix(profileURL, "https://github.com/") {
		return profileURL + ".png?s=512"
	}
	return ""
}
