package userdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type ProfileCache interface {
	GetTrades(ctx context.Context, uid string) ([]TradeRecord, bool, error)
	SaveTrades(ctx context.Context, uid string, v []TradeRecord) error
	InvalidateTrades(ctx context.Context, uid string) error

	GetValuation(ctx context.Context, uid string, now time.Time) ([]ValuationHistory, bool, error)
	SaveValuation(ctx context.Context, uid string, now time.Time, v []ValuationHistory) error
}

const (
	userTradeHistoryTTL          = time.Hour * 6
	portfolioValueChangeInterval = time.Hour
)

type userDBCache struct {
	r *redis.Client
}

func (_ userDBCache) tradesCacheKey(uid string) string { return fmt.Sprintf("trades::%s", uid) }
func (_ userDBCache) valuationCacheKey(uid string, now time.Time) string {
	return fmt.Sprintf("portfolioValuation::%s::%d", uid, now.Truncate(portfolioValueChangeInterval).Unix())
}

func (u userDBCache) GetTrades(ctx context.Context, uid string) ([]TradeRecord, bool, error) {
	var v cachedTradeHistory
	err := u.r.Get(ctx, u.tradesCacheKey(uid)).Scan(&v)
	return v, !errors.Is(err, redis.Nil), nonRedisNilErr(err)
}

func (u userDBCache) SaveTrades(ctx context.Context, uid string, v []TradeRecord) error {
	return u.r.Set(ctx, u.tradesCacheKey(uid), cachedTradeHistory(v), userTradeHistoryTTL).Err()
}

func (u userDBCache) InvalidateTrades(ctx context.Context, uid string) error {
	return u.r.Del(ctx, u.tradesCacheKey(uid)).Err()
}

func (u userDBCache) GetValuation(ctx context.Context, uid string, now time.Time) ([]ValuationHistory, bool, error) {
	var v cachedValuationHistory
	err := u.r.Get(ctx, u.valuationCacheKey(uid, now)).Scan(&v)
	return v, !errors.Is(err, redis.Nil), nonRedisNilErr(err)
}

func (u userDBCache) SaveValuation(ctx context.Context, uid string, now time.Time, v []ValuationHistory) error {
	return u.r.Set(ctx, u.valuationCacheKey(uid, now), cachedValuationHistory(v), portfolioValueChangeInterval).Err()
}

func nonRedisNilErr(err error) error {
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}

type cachedTradeHistory []TradeRecord

func (c cachedTradeHistory) MarshalBinary() (data []byte, err error) { return json.Marshal(c) }

func (c *cachedTradeHistory) UnmarshalBinary(data []byte) error { return json.Unmarshal(data, &c) }

type cachedValuationHistory []ValuationHistory

func (c cachedValuationHistory) MarshalBinary() (data []byte, err error) { return json.Marshal(c) }

func (c *cachedValuationHistory) UnmarshalBinary(data []byte) error { return json.Unmarshal(data, &c) }
