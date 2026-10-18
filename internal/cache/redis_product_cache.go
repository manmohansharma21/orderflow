// Package cache contains optional cache adapters.
package cache

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const productCacheVersionKey = "orderflow:products:cache-version"

// RedisProductCache uses a generation counter so invalidation is an O(1) INCR.
// Old generation keys become unreachable immediately and expire by TTL. This
// also avoids a stale in-flight reader re-populating a key used by new readers.
type RedisProductCache struct{ client *redis.Client }

func NewRedisProductCache(addr, password string, db int) *RedisProductCache {
	return &RedisProductCache{client: redis.NewClient(&redis.Options{
		Addr: addr, Password: password, DB: db,
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		PoolTimeout:  time.Second,
		MaxRetries:   0,
	})}
}

func (c *RedisProductCache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }
func (c *RedisProductCache) Close() error                   { return c.client.Close() }
func (c *RedisProductCache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}
func (c *RedisProductCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}
func (c *RedisProductCache) Version(ctx context.Context) (int64, error) {
	value, err := c.client.Get(ctx, productCacheVersionKey).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return version, nil
}
func (c *RedisProductCache) InvalidateProducts(ctx context.Context) error {
	return c.client.Incr(ctx, productCacheVersionKey).Err()
}
