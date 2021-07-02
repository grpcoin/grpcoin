package testutil

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func MockRedis(t *testing.T) *redis.Client {
	t.Helper()
	rs, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rs.Close() })
	return redis.NewClient(&redis.Options{
		Addr: rs.Addr(),
	})
}
