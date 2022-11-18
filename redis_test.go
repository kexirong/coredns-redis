package redis

import (
	"context"
	"testing"

	redisV8 "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func TestRedisGet(t *testing.T) {
	client := redisV8.NewUniversalClient(&redisV8.UniversalOptions{
		Addrs: []string{"192.168.0.165:6379"},
	})
	ret, err := client.HGet(ctx, "example.net.", "*").Result()
	t.Error(ret)
	t.Error(err)
}
