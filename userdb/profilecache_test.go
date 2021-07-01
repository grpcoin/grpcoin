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

package userdb

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
)

func TestTradeHistoryCaches(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
	c := UserDBCache{R: rc}

	_, ok, err := c.GetTrades(context.TODO(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("was not expecting value")
	}

	tr := []TradeRecord{
		{Date: time.Unix(100, 0),
			Ticker: "BTC",
			Action: 1,
			Size:   Amount{1, 1},
			Price:  Amount{2, 2}},
	}
	if err := c.SaveTrades(context.TODO(), "foo", tr); err != nil {
		t.Fatal(err)
	}
	got, ok, err := c.GetTrades(context.TODO(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("was expecting value")
	}
	if diff := cmp.Diff(tr, got); diff != "" {
		t.Fatal(diff)
	}

	if err := c.InvalidateTrades(context.TODO(), "foo"); err != nil {
		t.Fatal(err)
	}
	_, ok, err = c.GetTrades(context.TODO(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("was not expecting value post-invalidation")
	}
}

func TestPortfolioValidationHistoryCaches(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
	var c UserDBCache = UserDBCache{R: rc}
	now := time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC)
	_, ok, err := c.GetValuation(context.TODO(), "foo", now)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("was not expecting value")
	}

	pv := []ValuationHistory{
		{Date: time.Unix(1, 0), Value: Amount{1, 1}},
		{Date: time.Unix(2, 0), Value: Amount{2, 2}},
		{Date: time.Unix(3, 0), Value: Amount{3, 3}},
	}
	if err := c.SaveValuation(context.TODO(), "foo", now, pv); err != nil {
		t.Fatal(err)
	}

	// query too far ahead
	if _, ok, err := c.GetValuation(context.TODO(), "foo", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("was not expecting value bc queried too far ahead")
	}

	// query while cached data is valid
	v, ok, err := c.GetValuation(context.TODO(), "foo", now.Add(time.Minute*59))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("was expecting value")
	}
	if diff := cmp.Diff(pv, v); diff != "" {
		t.Fatal(diff)
	}
}
