package cache

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var once sync.Once

func InitRedisClient() error {
	once.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 10,
		})
	})
	ctx := context.Background()
	cmd := RedisClient.Ping(ctx)

	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}

func ClearCache() error {
	ctx := context.Background()
	_, err := RedisClient.FlushAll(ctx).Result()
	return err
}
