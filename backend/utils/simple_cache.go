package utils

import (
	"sync"
	"time"
)

type cacheItem struct {
	val    interface{}
	expire time.Time
}

// SimpleCache 是一个轻量的内存缓存，支持 TTL
type SimpleCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

// NewSimpleCache 创建缓存
func NewSimpleCache() *SimpleCache {
	c := &SimpleCache{items: make(map[string]cacheItem)}
	go c.janitor()
	return c
}

func (c *SimpleCache) janitor() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, it := range c.items {
			if !it.expire.IsZero() && it.expire.Before(now) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}

// Set 存储键值，ttl 为 0 表示不过期
func (c *SimpleCache) Set(key string, v interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	c.items[key] = cacheItem{val: v, expire: exp}
}

// Get 返回值和是否存在
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !it.expire.IsZero() && time.Now().After(it.expire) {
		// expired
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	return it.val, true
}

// DefaultCache 全局默认缓存实例
var DefaultCache = NewSimpleCache()
