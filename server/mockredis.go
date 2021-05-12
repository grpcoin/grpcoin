package main

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type hook struct{}

func (h *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	cmd.SetErr(redis.Nil)
	return ctx, nil
}
func (h *hook) AfterProcess(_ context.Context, _ redis.Cmder) error {
	return nil
}
func (h *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	for _, cmd := range cmds {
		cmd.SetErr(redis.Nil)
	}
	return ctx, nil
}
func (h *hook) AfterProcessPipeline(_ context.Context, _ []redis.Cmder) error {
	return nil
}

func dummyRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		MaxRetries: -2,
	})
	rdb.AddHook(&hook{})
	return rdb
}
