package community

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test JSONHelper
func TestJSONHelper_Marshal(t *testing.T) {
	helper := NewJSONHelper()

	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	result, err := helper.Marshal(data)

	assert.NoError(t, err)
	assert.Contains(t, string(result), "key1")
	assert.Contains(t, string(result), "value1")
}

func TestJSONHelper_Unmarshal(t *testing.T) {
	helper := NewJSONHelper()

	jsonData := []byte(`{"key1":"value1","key2":123}`)
	var result map[string]interface{}

	err := helper.Unmarshal(jsonData, &result)

	assert.NoError(t, err)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, float64(123), result["key2"]) // JSON numbers are float64
}

func TestJSONHelper_MarshalToString(t *testing.T) {
	helper := NewJSONHelper()

	data := map[string]string{"test": "value"}

	result, err := helper.MarshalToString(data)

	assert.NoError(t, err)
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "value")
}

func TestJSONHelper_UnmarshalFromString(t *testing.T) {
	helper := NewJSONHelper()

	jsonStr := `{"test":"value"}`
	var result map[string]string

	err := helper.UnmarshalFromString(jsonStr, &result)

	assert.NoError(t, err)
	assert.Equal(t, "value", result["test"])
}

// Test CacheHelper
func TestCacheHelper_Get_Success(t *testing.T) {
	mockRedis := new(MockRedisClient)
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(mockRedis, jsonHelper)

	ctx := context.Background()
	key := "test_key"
	cachedData := `{"test":"value"}`
	var result map[string]string

	mockRedis.On("Get", ctx, key).Return(cachedData, nil)

	err := cacheHelper.Get(ctx, key, &result)

	assert.NoError(t, err)
	assert.Equal(t, "value", result["test"])
	mockRedis.AssertExpectations(t)
}

func TestCacheHelper_Get_CacheMiss(t *testing.T) {
	mockRedis := new(MockRedisClient)
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(mockRedis, jsonHelper)

	ctx := context.Background()
	key := "test_key"
	var result map[string]string

	mockRedis.On("Get", ctx, key).Return("", nil)

	err := cacheHelper.Get(ctx, key, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache miss")
	mockRedis.AssertExpectations(t)
}

func TestCacheHelper_Set_Success(t *testing.T) {
	mockRedis := new(MockRedisClient)
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(mockRedis, jsonHelper)

	ctx := context.Background()
	key := "test_key"
	data := map[string]string{"test": "value"}
	expiration := time.Minute

	mockRedis.On("Set", ctx, key, mock.AnythingOfType("string"), expiration).Return(nil)

	err := cacheHelper.Set(ctx, key, data, expiration)

	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestCacheHelper_Delete_Success(t *testing.T) {
	mockRedis := new(MockRedisClient)
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(mockRedis, jsonHelper)

	ctx := context.Background()
	keys := []string{"key1", "key2"}

	mockRedis.On("Del", ctx, keys).Return(nil)

	err := cacheHelper.Delete(ctx, keys...)

	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestCacheHelper_GetOrSet_CacheHit(t *testing.T) {
	mockRedis := new(MockRedisClient)
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(mockRedis, jsonHelper)

	ctx := context.Background()
	key := "test_key"
	cachedData := `{"test":"value"}`
	var result map[string]string

	mockRedis.On("Get", ctx, key).Return(cachedData, nil)

	fetchCalled := false
	fetchFunc := func() (any, error) {
		fetchCalled = true
		return map[string]string{"test": "new_value"}, nil
	}

	err := cacheHelper.GetOrSet(ctx, key, &result, time.Minute, fetchFunc)

	assert.NoError(t, err)
	assert.Equal(t, "value", result["test"])
	assert.False(t, fetchCalled) // Should not call fetch function on cache hit
	mockRedis.AssertExpectations(t)
}

// Test ConfigHelper
func TestConfigHelper_ValidateTimeWindow(t *testing.T) {
	helper := NewConfigHelper()

	tests := []struct {
		timeWindow string
		expected   bool
	}{
		{"hourly", true},
		{"daily", true},
		{"weekly", true},
		{"monthly", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.timeWindow, func(t *testing.T) {
			result := helper.ValidateTimeWindow(tt.timeWindow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigHelper_ValidateAlgorithm(t *testing.T) {
	helper := NewConfigHelper()

	tests := []struct {
		algorithm string
		expected  bool
	}{
		{"collaborative", true},
		{"content_based", true},
		{"trending", true},
		{"popularity", true},
		{"category", true},
		{"hybrid", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			result := helper.ValidateAlgorithm(tt.algorithm)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigHelper_GetTimeRange(t *testing.T) {
	helper := NewConfigHelper()

	tests := []struct {
		timeWindow string
		shouldErr  bool
	}{
		{"hourly", false},
		{"daily", false},
		{"weekly", false},
		{"monthly", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.timeWindow, func(t *testing.T) {
			result, err := helper.GetTimeRange(tt.timeWindow)
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidTimeWindow, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.Before(time.Now()))
			}
		})
	}
}

// Test ValidationHelper
func TestValidationHelper_ValidateUserID(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		userID   string
		expected error
	}{
		{"user-123", nil},
		{"", ErrInvalidUserID},
	}

	for _, tt := range tests {
		t.Run(tt.userID, func(t *testing.T) {
			err := helper.ValidateUserID(tt.userID)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestValidationHelper_ValidateBookmarkID(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		bookmarkID uint
		expected   error
	}{
		{1, nil},
		{123, nil},
		{0, ErrInvalidBookmarkID},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := helper.ValidateBookmarkID(tt.bookmarkID)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestValidationHelper_ValidateActionType(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		actionType string
		expected   error
	}{
		{"view", nil},
		{"click", nil},
		{"save", nil},
		{"share", nil},
		{"like", nil},
		{"dismiss", nil},
		{"report", nil},
		{"follow", nil},
		{"unfollow", nil},
		{"invalid", ErrInvalidActionType},
		{"", ErrInvalidActionType},
	}

	for _, tt := range tests {
		t.Run(tt.actionType, func(t *testing.T) {
			err := helper.ValidateActionType(tt.actionType)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestValidationHelper_ValidateScore(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		score    float64
		expected error
	}{
		{0.0, nil},
		{0.5, nil},
		{1.0, nil},
		{-0.1, ErrInvalidScore},
		{1.1, ErrInvalidScore},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := helper.ValidateScore(tt.score)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestValidationHelper_ValidateLimit(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		limit    int
		expected int
	}{
		{10, 10},
		{50, 50},
		{100, 100},
		{0, 20},   // Default
		{-5, 20},  // Default
		{150, 20}, // Default (over max)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := helper.ValidateLimit(tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}
