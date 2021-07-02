// Package tradecounters is a library for counting trade events
// happening on the platform.
package tradecounters

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpcoin/grpcoin/testutil"
)

func TestTradeCounters(t *testing.T) {
	r := testutil.MockRedis(t)
	ctx := context.Background()
	counter := TradeCounter{DB: r}

	now := time.Date(2050, time.January, 10, 10, 30, 20, 0, time.UTC)
	minusHr := func(i int) time.Time { return now.Add(-time.Duration(i) * time.Hour) }
	minusMin := func(i int) time.Time { return now.Add(-time.Duration(i) * time.Minute) }

	if err := counter.IncrTrades(ctx, minusHr(30), 100_000); err != nil {
		t.Fatal(err)
	}
	if err := counter.IncrTrades(ctx, minusHr(25), 10_000); err != nil {
		t.Fatal(err)
	}
	if err := counter.IncrTrades(ctx, minusHr(10), 15_000); err != nil {
		t.Fatal(err)
	}
	if err := counter.IncrTrades(ctx, minusHr(1), 3_000); err != nil {
		t.Fatal(err)
	}
	if err := counter.IncrTrades(ctx, minusMin(1), 50); err != nil {
		t.Fatal(err)
	}
	if err := counter.IncrTrades(ctx, minusHr(-1), 1); err != nil {
		t.Fatal(err)
	}

	dailyTrades, err := counter.PastDayTradeCounts(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if dailyTrades != 3 {
		t.Errorf("Expected 3 trades, got %d", dailyTrades)
	}

	tradeVolume, err := counter.PastDayTradeVolume(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	expectedTradeVolume := float32(18050.0)
	if cmp.Equal(expectedTradeVolume, tradeVolume, cmpopts.EquateApprox(0, 0.0001)) {
		t.Errorf("Expected %f trade volume, got %f", expectedTradeVolume, tradeVolume)
	}
}
