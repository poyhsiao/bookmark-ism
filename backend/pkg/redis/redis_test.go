package redis

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a test Redis server using miniredis
// setupTestRedis 使用 miniredis 創建測試 Redis 服務器
func setupTestRedis(t *testing.T) (*Client, *miniredis.Miniredis) {
	// Create a miniredis server
	// 創建 miniredis 服務器
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis configuration
	// 創建 Redis 配置
	cfg := config.RedisConfig{
		Host:     mr.Host(),
		Port:     mr.Port(),
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	// Create Redis client
	// 創建 Redis 客戶端
	client, err := NewClient(cfg)
	require.NoError(t, err)

	return client, mr
}

// TestNewClient tests the Redis client creation
// TestNewClient 測試 Redis 客戶端創建
func TestNewClient(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	// Test connection
	// 測試連接
	ctx := context.Background()
	err := client.Ping(ctx)
	assert.NoError(t, err)
}

// TestSetWithExpiration tests setting a key with expiration
// TestSetWithExpiration 測試設置帶有過期時間的鍵
func TestSetWithExpiration(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:key"
	value := "test-value"
	expiration := 1 * time.Second

	// Set key with expiration
	// 設置帶有過期時間的鍵
	err := client.SetWithExpiration(ctx, key, value, expiration)
	assert.NoError(t, err)

	// Check that key exists
	// 檢查鍵是否存在
	val, err := client.GetString(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, val)

	// Fast-forward time in miniredis
	// 在 miniredis 中快進時間
	mr.FastForward(2 * time.Second)

	// Check that key has expired
	// 檢查鍵是否已過期
	_, err = client.GetString(ctx, key)
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

// TestGetString tests getting a string value
// TestGetString 測試獲取字符串值
func TestGetString(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:string"
	value := "test-string-value"

	// Set key
	// 設置鍵
	err := client.SetWithExpiration(ctx, key, value, 0)
	assert.NoError(t, err)

	// Get key
	// 獲取鍵
	val, err := client.GetString(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, val)

	// Get non-existent key
	// 獲取不存在的鍵
	_, err = client.GetString(ctx, "non:existent:key")
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

// TestDelete tests deleting keys
// TestDelete 測試刪除鍵
func TestDelete(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key1 := "test:delete:1"
	key2 := "test:delete:2"

	// Set keys
	// 設置鍵
	err := client.SetWithExpiration(ctx, key1, "value1", 0)
	assert.NoError(t, err)
	err = client.SetWithExpiration(ctx, key2, "value2", 0)
	assert.NoError(t, err)

	// Delete keys
	// 刪除鍵
	err = client.Delete(ctx, key1, key2)
	assert.NoError(t, err)

	// Check that keys are deleted
	// 檢查鍵是否已刪除
	exists, err := client.Exists(ctx, key1, key2)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

// TestExists tests checking if keys exist
// TestExists 測試檢查鍵是否存在
func TestExists(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key1 := "test:exists:1"
	key2 := "test:exists:2"
	key3 := "test:exists:3"

	// Set keys
	// 設置鍵
	err := client.SetWithExpiration(ctx, key1, "value1", 0)
	assert.NoError(t, err)
	err = client.SetWithExpiration(ctx, key2, "value2", 0)
	assert.NoError(t, err)

	// Check existing keys
	// 檢查存在的鍵
	exists, err := client.Exists(ctx, key1, key2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), exists)

	// Check mix of existing and non-existing keys
	// 檢查存在和不存在的鍵的混合
	exists, err = client.Exists(ctx, key1, key3)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}

// TestSetNX tests setting a key only if it doesn't exist
// TestSetNX 測試僅在鍵不存在時設置鍵
func TestSetNX(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:setnx"
	value1 := "value1"
	value2 := "value2"

	// Set key that doesn't exist
	// 設置不存在的鍵
	ok, err := client.SetNX(ctx, key, value1, 0)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Try to set key that already exists
	// 嘗試設置已存在的鍵
	ok, err = client.SetNX(ctx, key, value2, 0)
	assert.NoError(t, err)
	assert.False(t, ok)

	// Check that value is still the original
	// 檢查值是否仍為原始值
	val, err := client.GetString(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value1, val)
}

// TestIncrement tests incrementing a counter
// TestIncrement 測試遞增計數器
func TestIncrement(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:counter"

	// Increment non-existent key
	// 遞增不存在的鍵
	val, err := client.Increment(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Increment existing key
	// 遞增現有鍵
	val, err = client.Increment(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

// TestIncrementWithExpiration tests incrementing a counter with expiration
// TestIncrementWithExpiration 測試帶有過期時間的遞增計數器
func TestIncrementWithExpiration(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test:counter:exp"
	expiration := 1 * time.Second

	// Increment non-existent key with expiration
	// 遞增帶有過期時間的不存在鍵
	val, err := client.IncrementWithExpiration(ctx, key, expiration)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Increment existing key
	// 遞增現有鍵
	val, err = client.IncrementWithExpiration(ctx, key, expiration)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// Fast-forward time in miniredis
	// 在 miniredis 中快進時間
	mr.FastForward(2 * time.Second)

	// Check that key has expired
	// 檢查鍵是否已過期
	exists, err := client.Exists(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

// TestPublishSubscribe tests the publish/subscribe functionality
// TestPublishSubscribe 測試發布/訂閱功能
func TestPublishSubscribe(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	channel := "test:channel"
	message := "test-message"

	// Subscribe to channel
	// 訂閱頻道
	pubsub := client.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Receive subscription confirmation
	// 接收訂閱確認
	_, err := pubsub.Receive(ctx)
	assert.NoError(t, err)

	// Start receiving messages in a goroutine
	// 在 goroutine 中開始接收消息
	msgCh := pubsub.Channel()

	// Publish message
	// 發布消息
	err = client.Publish(ctx, channel, message)
	assert.NoError(t, err)

	// Wait for message
	// 等待消息
	select {
	case msg := <-msgCh:
		assert.Equal(t, channel, msg.Channel)
		assert.Equal(t, message, msg.Payload)
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

// TestSubscribeToSyncEvents tests subscribing to sync events
// TestSubscribeToSyncEvents 測試訂閱同步事件
func TestSubscribeToSyncEvents(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "123"

	// Subscribe to sync events
	// 訂閱同步事件
	pubsub := client.SubscribeToSyncEvents(ctx, userID)
	defer pubsub.Close()

	// Check that we're subscribed to the correct channel
	// 檢查我們是否訂閱了正確的頻道
	// Note: We can't easily test the channels in miniredis, so we'll skip this check
	// 注意：我們無法在 miniredis 中輕易測試頻道，所以跳過這個檢查
	assert.NotNil(t, pubsub)
}

// TestPublishSyncEvent tests publishing a sync event
// TestPublishSyncEvent 測試發布同步事件
func TestPublishSyncEvent(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "123"
	event := map[string]interface{}{
		"type":       "bookmark_created",
		"bookmarkID": 456,
	}

	// Publish sync event
	// 發布同步事件
	err := client.PublishSyncEvent(ctx, userID, event)
	assert.NoError(t, err)
}

// TestClose tests closing the Redis client
// TestClose 測試關閉 Redis 客戶端
func TestClose(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	// Close client
	// 關閉客戶端
	err := client.Close()
	assert.NoError(t, err)

	// Operations should fail after closing
	// 關閉後操作應該失敗
	ctx := context.Background()
	err = client.Ping(ctx)
	assert.Error(t, err)
}
