package serverutil

import (
	"context"
	"log"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConnectRedis establishes a connection to redis, or initializes an
// in-memory implementation if ip is empty.
func ConnectRedis(ctx context.Context, redisIP string) (rc *redis.Client, close func(), err error) {
	if redisIP == "" {
		s, err := miniredis.Run()
		if err != nil {
		}
		rc = redis.NewClient(&redis.Options{Addr: s.Addr()})
		close = s.Close
	} else {
		rc = redis.NewClient(&redis.Options{Addr: redisIP + ":6379"})
		close = func() {
		}
	}
	if err := rc.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping failed", zap.Error(err))
	}
	return rc, close, nil

}
