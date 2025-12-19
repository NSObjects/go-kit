// Package cache provides Redis cache abstraction.
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache provides a simple cache interface.
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// RedisCache implements Cache using Redis.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache.
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

func (c *RedisCache) key(k string) string {
	if c.prefix == "" {
		return k
	}
	return c.prefix + ":" + k
}

// Get retrieves a value from cache.
func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := c.client.Get(ctx, c.key(key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set stores a value in cache.
func (c *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(key), data, expiration).Err()
}

// Delete removes a value from cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

// Exists checks if a key exists in cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, c.key(key)).Result()
	return result > 0, err
}
