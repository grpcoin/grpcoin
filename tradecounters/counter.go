// Package tradecounters is a library for counting trade events
// happening on the platform.
package tradecounters

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type TradeCounter struct {
	DB *redis.Client
}

// IncrTrades increments trade count and volume by specified amounts.
func (tc TradeCounter) IncrTrades(ctx context.Context, now time.Time, volume float64) error {
	p := tc.DB.Pipeline()
	defer p.Close()
	p.Incr(ctx, keyHourlyTradeCount(now))
	p.ExpireAt(ctx, keyHourlyTradeCount(now), keyExpiration(now))
	p.IncrByFloat(ctx, keyHourlyTradeVolume(now), volume)
	p.ExpireAt(ctx, keyHourlyTradeVolume(now), keyExpiration(now))
	cmds, err := p.Exec(ctx)
	_ = cmds
	return err
}

func (tc TradeCounter) PastDayTradeCounts(ctx context.Context, now time.Time) (int, error) {
	keys := genCacheKeys(now, 24, keyHourlyTradeCount)
	res, err := tc.DB.MGet(ctx, keys...).Result()
	var total int
	// TODO it sucks to parse results like this, see https://github.com/go-redis/redis/issues/1810
	for _, c := range res {
		if v, ok := c.(string); ok {
			vv, _ := strconv.Atoi(v)
			total += vv
		}

	}
	return total, err
}

func (tc TradeCounter) PastDayTradeVolume(ctx context.Context, now time.Time) (float64, error) {
	res, err := tc.DB.MGet(ctx, genCacheKeys(now, 24, keyHourlyTradeVolume)...).Result()
	var total float64
	for _, c := range res {
		if v, ok := c.(string); ok {
			vv, _ := strconv.ParseFloat(v, 64)
			total += vv
		}
	}
	return total, err
}

func keyExpiration(t time.Time) time.Time {
	return t.Truncate(time.Hour).Add(time.Hour * 25)
}

// genCacheKeys generates cache keys looking past from the specified now time
// for the specified number of hours using the given cache key function.
func genCacheKeys(now time.Time, nHours int, keyFunc func(time.Time) string) []string {
	keys := make([]string, 0, nHours)
	for i := 0; i < nHours; i++ {
		keys = append(keys, keyFunc(now.Add(-time.Hour*time.Duration(i))))
	}
	return keys
}

func keyHourlyTradeVolume(t time.Time) string {
	return fmt.Sprintf("tradevolume_hr::%s", t.Truncate(time.Hour).Format(time.RFC3339))
}

func keyHourlyTradeCount(t time.Time) string {
	return fmt.Sprintf("trades_hr::%s", t.Truncate(time.Hour).Format(time.RFC3339))
}
