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
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type AuthToken interface {
	V() string
}

type AccountCache struct {
	cache *redis.Client
}

func (a *AccountCache) Set(t AuthToken, uid string) error {
	return a.cache.Set(context.TODO(), a.cacheKey(t), uid, time.Hour*2).Err()
}

func (a *AccountCache) Get(t AuthToken) (string, bool, error) {
	s, err := a.cache.Get(context.TODO(), a.cacheKey(t)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	return s, true, err
}

func (a *AccountCache) cacheKey(t AuthToken) string {
	h := sha256.New()
	h.Write([]byte(t.V()))
	return fmt.Sprintf("token_%x", h.Sum(nil))
}
