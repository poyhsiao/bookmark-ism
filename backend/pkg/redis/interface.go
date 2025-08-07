package redis

import (
	"context"
	"time"
)

// RedisInterface defines the Redis operations needed by the application
// RedisInterface 定義應用程序所需的 Redis 操作
type RedisInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
}

// Ensure Client implements RedisInterface
// 確保 Client 實現 RedisInterface
var _ RedisInterface = (*Client)(nil)

// Set implements RedisInterface
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Get implements RedisInterface
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Del implements RedisInterface
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Ping implements RedisInterface
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
