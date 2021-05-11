package main

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type AuthToken interface {
	V() string
}

type GitHubAccessToken string

func (g GitHubAccessToken) V() string { return string(g) }

type AccountCache struct {
	cache *redis.Client
}

func (a *AccountCache) Set(t AuthToken, uid string) error {
	return a.cache.Set(context.TODO(), a.cacheKey(t), uid, 0).Err()
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
