package community

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// JSONHelper provides JSON marshaling/unmarshaling utilities
type JSONHelper struct{}

// NewJSONHelper creates a new JSON helper
func NewJSONHelper() *JSONHelper {
	return &JSONHelper{}
}

// Marshal marshals data to JSON bytes
func (h *JSONHelper) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

// Unmarshal unmarshals JSON bytes to data
func (h *JSONHelper) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// MarshalToString marshals data to JSON string
func (h *JSONHelper) MarshalToString(data any) (string, error) {
	bytes, err := h.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// UnmarshalFromString unmarshals JSON string to data
func (h *JSONHelper) UnmarshalFromString(jsonStr string, v any) error {
	return h.Unmarshal([]byte(jsonStr), v)
}

// CacheHelper provides caching utilities with JSON serialization
type CacheHelper struct {
	redis      RedisClient
	jsonHelper *JSONHelper
}

// NewCacheHelper creates a new cache helper
func NewCacheHelper(redis RedisClient, jsonHelper *JSONHelper) *CacheHelper {
	return &CacheHelper{
		redis:      redis,
		jsonHelper: jsonHelper,
	}
}

// Get retrieves and unmarshals cached data
func (h *CacheHelper) Get(ctx context.Context, key string, dest any) error {
	cached, err := h.redis.Get(ctx, key)
	if err != nil {
		return err
	}
	if cached == "" {
		return fmt.Errorf("cache miss for key: %s", key)
	}
	return h.jsonHelper.UnmarshalFromString(cached, dest)
}

// Set marshals and caches data
func (h *CacheHelper) Set(ctx context.Context, key string, data any, expiration time.Duration) error {
	jsonStr, err := h.jsonHelper.MarshalToString(data)
	if err != nil {
		return err
	}
	return h.redis.Set(ctx, key, jsonStr, expiration)
}

// Delete removes cached data
func (h *CacheHelper) Delete(ctx context.Context, keys ...string) error {
	return h.redis.Del(ctx, keys...)
}

// GetOrSet retrieves cached data or sets it if not found
func (h *CacheHelper) GetOrSet(ctx context.Context, key string, dest any, expiration time.Duration, fetchFunc func() (any, error)) error {
	// Try to get from cache first
	err := h.Get(ctx, key, dest)
	if err == nil {
		return nil // Cache hit
	}

	// Cache miss, fetch data
	data, err := fetchFunc()
	if err != nil {
		return err
	}

	// Set in cache
	if err := h.Set(ctx, key, data, expiration); err != nil {
		// Log error but don't fail the request
		return nil
	}

	// Copy data to destination
	jsonStr, err := h.jsonHelper.MarshalToString(data)
	if err != nil {
		return err
	}
	return h.jsonHelper.UnmarshalFromString(jsonStr, dest)
}

// ConfigHelper provides configuration utilities
type ConfigHelper struct{}

// NewConfigHelper creates a new config helper
func NewConfigHelper() *ConfigHelper {
	return &ConfigHelper{}
}

// ValidateTimeWindow validates time window values
func (h *ConfigHelper) ValidateTimeWindow(timeWindow string) bool {
	validWindows := map[string]bool{
		"hourly": true, "daily": true, "weekly": true, "monthly": true,
	}
	return validWindows[timeWindow]
}

// ValidateAlgorithm validates recommendation algorithm values
func (h *ConfigHelper) ValidateAlgorithm(algorithm string) bool {
	validAlgorithms := map[string]bool{
		"collaborative": true, "content_based": true, "trending": true,
		"popularity": true, "category": true, "hybrid": true,
	}
	return validAlgorithms[algorithm]
}

// GetTimeRange calculates time range for given window
func (h *ConfigHelper) GetTimeRange(timeWindow string) (time.Time, error) {
	now := time.Now()
	switch timeWindow {
	case "hourly":
		return now.Add(-1 * time.Hour), nil
	case "daily":
		return now.Add(-24 * time.Hour), nil
	case "weekly":
		return now.Add(-7 * 24 * time.Hour), nil
	case "monthly":
		return now.Add(-30 * 24 * time.Hour), nil
	default:
		return time.Time{}, ErrInvalidTimeWindow
	}
}

// ValidationHelper provides validation utilities
type ValidationHelper struct{}

// NewValidationHelper creates a new validation helper
func NewValidationHelper() *ValidationHelper {
	return &ValidationHelper{}
}

// ValidateUserID validates user ID
func (h *ValidationHelper) ValidateUserID(userID string) error {
	if userID == "" {
		return ErrInvalidUserID
	}
	return nil
}

// ValidateBookmarkID validates bookmark ID
func (h *ValidationHelper) ValidateBookmarkID(bookmarkID uint) error {
	if bookmarkID == 0 {
		return ErrInvalidBookmarkID
	}
	return nil
}

// ValidateActionType validates action type
func (h *ValidationHelper) ValidateActionType(actionType string) error {
	if actionType == "" {
		return ErrInvalidActionType
	}
	validActions := map[string]bool{
		"view": true, "click": true, "save": true, "share": true, "like": true,
		"dismiss": true, "report": true, "follow": true, "unfollow": true,
	}
	if !validActions[actionType] {
		return ErrInvalidActionType
	}
	return nil
}

// ValidateScore validates recommendation score
func (h *ValidationHelper) ValidateScore(score float64) error {
	if score < 0 || score > 1 {
		return ErrInvalidScore
	}
	return nil
}

// ValidateLimit validates pagination limit
func (h *ValidationHelper) ValidateLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 20 // Default limit
	}
	return limit
}
