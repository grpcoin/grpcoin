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

package serverutil

import (
	"context"
	"log"
	"net"

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
		rc = redis.NewClient(&redis.Options{Addr: net.JoinHostPort(redisIP, "6379")})
		close = func() {}
	}
	if err := rc.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping failed", zap.Error(err))
	}
	return rc, close, nil

}
