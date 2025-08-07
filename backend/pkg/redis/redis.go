package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/go-redis/redis/v8"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	*redis.Client
	pubsub *redis.PubSub
}

// NewClient creates a new Redis client with connection pooling
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.RedisConnectionTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		Client: rdb,
	}, nil
}

// Subscribe subscribes to Redis channels for pub/sub messaging
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	c.pubsub = c.Client.Subscribe(ctx, channels...)
	return c.pubsub
}

// PSubscribe subscribes to Redis channels using patterns
func (c *Client) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	c.pubsub = c.Client.PSubscribe(ctx, patterns...)
	return c.pubsub
}

// Publish publishes a message to a Redis channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.Client.Publish(ctx, channel, message).Err()
}

// PublishJSON publishes a JSON message to a Redis channel
func (c *Client) PublishJSON(ctx context.Context, channel string, message interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	return c.Client.Publish(ctx, channel, jsonData).Err()
}

// SubscribeToSyncEvents subscribes to bookmark sync events
func (c *Client) SubscribeToSyncEvents(ctx context.Context, userID string) *redis.PubSub {
	channel := fmt.Sprintf("sync:user:%s", userID)
	return c.Subscribe(ctx, channel)
}

// PublishSyncEvent publishes a sync event for a user
func (c *Client) PublishSyncEvent(ctx context.Context, userID string, event interface{}) error {
	channel := fmt.Sprintf("sync:user:%s", userID)
	return c.PublishJSON(ctx, channel, event)
}

// SubscribeToNotifications subscribes to user notifications
func (c *Client) SubscribeToNotifications(ctx context.Context, userID string) *redis.PubSub {
	channel := fmt.Sprintf("notifications:user:%s", userID)
	return c.Subscribe(ctx, channel)
}

// PublishNotification publishes a notification for a user
func (c *Client) PublishNotification(ctx context.Context, userID string, notification interface{}) error {
	channel := fmt.Sprintf("notifications:user:%s", userID)
	return c.PublishJSON(ctx, channel, notification)
}

// Close closes the Redis connection and any active subscriptions
func (c *Client) Close() error {
	if c.pubsub != nil {
		if err := c.pubsub.Close(); err != nil {
			return fmt.Errorf("failed to close pubsub: %w", err)
		}
	}
	return c.Client.Close()
}

// SetWithExpiration sets a key-value pair with expiration
func (c *Client) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// GetString gets a string value by key
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Delete deletes keys
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiration).Result()
}

// Increment increments a counter
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// IncrementWithExpiration increments a counter and sets expiration if key is new
func (c *Client) IncrementWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := c.Client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
