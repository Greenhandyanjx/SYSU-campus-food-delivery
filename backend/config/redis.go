package config

import (
	"backend/global"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// InitRedis 初始化 Redis 客户端，配置从 config.yaml 中读取（可选），默认 localhost:6379
func InitRedis() error {
	// 尝试从配置读取 redis.host/port/password
	host := "localhost"
	port := "6379"
	pass := ""
	if viper.IsSet("redis.host") {
		host = viper.GetString("redis.host")
	}
	if viper.IsSet("redis.port") {
		port = viper.GetString("redis.port")
	}
	if viper.IsSet("redis.password") {
		pass = viper.GetString("redis.password")
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})

	// 简单的 ping 测试
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	global.RedisClient = client
	return nil
}
