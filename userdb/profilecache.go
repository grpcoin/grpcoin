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

type UserDBCache struct {
	R *redis.Client
}

func (_ UserDBCache) tradesCacheKey(uid string) string { return fmt.Sprintf("trades::%s", uid) }
func (_ UserDBCache) valuationCacheKey(uid string, now time.Time) string {
	return fmt.Sprintf("portfolioValuation::%s::%d", uid, now.Truncate(portfolioValueChangeInterval).Unix())
}

func (u UserDBCache) GetTrades(ctx context.Context, uid string) ([]TradeRecord, bool, error) {
	var v cachedTradeHistory
	err := u.R.Get(ctx, u.tradesCacheKey(uid)).Scan(&v)
	return v, !errors.Is(err, redis.Nil), nonRedisNilErr(err)
}

func (u UserDBCache) SaveTrades(ctx context.Context, uid string, v []TradeRecord) error {
	return u.R.Set(ctx, u.tradesCacheKey(uid), cachedTradeHistory(v), userTradeHistoryTTL).Err()
}

func (u UserDBCache) InvalidateTrades(ctx context.Context, uid string) error {
	return u.R.Del(ctx, u.tradesCacheKey(uid)).Err()
}

func (u UserDBCache) GetValuation(ctx context.Context, uid string, now time.Time) ([]ValuationHistory, bool, error) {
	var v cachedValuationHistory
	err := u.R.Get(ctx, u.valuationCacheKey(uid, now)).Scan(&v)
	return v, !errors.Is(err, redis.Nil), nonRedisNilErr(err)
}

func (u UserDBCache) SaveValuation(ctx context.Context, uid string, now time.Time, v []ValuationHistory) error {
	return u.R.Set(ctx, u.valuationCacheKey(uid, now), cachedValuationHistory(v), portfolioValueChangeInterval).Err()
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
