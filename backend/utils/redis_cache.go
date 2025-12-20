package utils

import (
	"context"
	"encoding/json"
	"fmt"
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
	fmt.Println("Getting key: "+key);
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
	fmt.Println("Setting key: "+key);
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

func ScanAndDeleteKeys(ctx context.Context, pattern string) error {
    c := redisClient()
    if c == nil {
        return nil
    }
    fmt.Println("Scanning and deleting keys with pattern: ", pattern)
    iter := c.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        key := iter.Val()
		fmt.Println(key);
        if err := c.Del(ctx, key).Err(); err != nil {
            return err
        }
		
    }
    if err := iter.Err(); err != nil {
        return err
    }
    return nil
}
