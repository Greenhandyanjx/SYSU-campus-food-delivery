package utils

import (
	"context"
	"encoding/json"
	"time"

	"backend/global"

	"github.com/go-redis/redis/v8"
)

// helper to get redis client
func redisClient() *redis.Client {
	if global.RedisClient == nil {
		return nil
	}
	if c, ok := global.RedisClient.(*redis.Client); ok {
		return c
	}
	return nil
}

// GetJSON tries to GET key and unmarshal into dest. Returns found bool.
func GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	c := redisClient()
	if c == nil {
		return false, nil
	}
	s, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(s), dest); err != nil {
		return false, err
	}
	return true, nil
}

// SetJSON marshals value and sets key with ttl
func SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c := redisClient()
	if c == nil {
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, b, ttl).Err()
}

// Del removes key
func Del(ctx context.Context, key string) error {
	c := redisClient()
	if c == nil {
		return nil
	}
	return c.Del(ctx, key).Err()
}
