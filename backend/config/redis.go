package config

import (
	"backend/global"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// InitRedis initializes a Redis client with sensible defaults (localhost:6379).
func InitRedis() error {
	// Use defaults; you can change host/port as needed
	opt := &redis.Options{
		Addr:        "localhost:6379",
		Password:    "", // no password
		DB:          0,
		DialTimeout: 5 * time.Second,
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	// store client in global as the concrete type (so other packages can type assert)
	global.RedisClient = client
	return nil
}
