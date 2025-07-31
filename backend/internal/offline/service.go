package offline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bookmark-sync-service/backend/pkg/database"

	"gorm.io/gorm"
)

// Common errors
var (
	ErrBookmarkNotCached  = errors.New("bookmark not found in cache")
	ErrInvalidChangeType  = errors.New("invalid change type")
	ErrConflictResolution = errors.New("failed to resolve conflict")
	ErrKeyNotFound        = errors.New("key not found")
)

// RedisClientInterface defines the interface for Redis operations
type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	IncrementWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Publish(ctx context.Context, channel string, message interface{}) error
	Close() error
}

// OfflineChange represents a change made while offline
type OfflineChange struct {
	ID         string    `json:"id"`
	UserID     uint      `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	Type       string    `json:"type"` // bookmark_create, bookmark_update, bookmark_delete, collection_create, etc.
	ResourceID string    `json:"resource_id"`
	Data       string    `json:"data"` // JSON data of the change
	Timestamp  time.Time `json:"timestamp"`
	Applied    bool      `json:"applied"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	CachedBookmarksCount int       `json:"cached_bookmarks"`
	QueuedChangesCount   int       `json:"queued_changes"`
	LastSync             time.Time `json:"last_sync"`
	CacheSize            int64     `json:"cache_size"`
}

// Service handles offline functionality
type Service struct {
	db          *gorm.DB
	redisClient RedisClientInterface
}

// NewService creates a new offline service
func NewService(db *gorm.DB, redisClient RedisClientInterface) *Service {
	return &Service{
		db:          db,
		redisClient: redisClient,
	}
}

// CacheBookmark stores a bookmark in the offline cache
func (s *Service) CacheBookmark(ctx context.Context, bookmark *database.Bookmark) error {
	key := fmt.Sprintf("offline:bookmark:%d:%d", bookmark.UserID, bookmark.ID)

	bookmarkJSON, err := json.Marshal(bookmark)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmark: %w", err)
	}

	// Cache for 24 hours
	err = s.redisClient.Set(ctx, key, string(bookmarkJSON), time.Hour*24)
	if err != nil {
		return fmt.Errorf("failed to cache bookmark: %w", err)
	}

	return nil
}

// GetCachedBookmark retrieves a bookmark from the offline cache
func (s *Service) GetCachedBookmark(ctx context.Context, userID, bookmarkID uint) (*database.Bookmark, error) {
	key := fmt.Sprintf("offline:bookmark:%d:%d", userID, bookmarkID)

	bookmarkJSON, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" || err.Error() == "key not found" {
			return nil, ErrBookmarkNotCached
		}
		return nil, fmt.Errorf("failed to get cached bookmark: %w", err)
	}

	var bookmark database.Bookmark
	err = json.Unmarshal([]byte(bookmarkJSON), &bookmark)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookmark: %w", err)
	}

	return &bookmark, nil
}

// CacheBookmarks stores multiple bookmarks in the offline cache
func (s *Service) CacheBookmarks(ctx context.Context, bookmarks []database.Bookmark) error {
	for _, bookmark := range bookmarks {
		if err := s.CacheBookmark(ctx, &bookmark); err != nil {
			return fmt.Errorf("failed to cache bookmark %d: %w", bookmark.ID, err)
		}
	}
	return nil
}

// GetCachedBookmarksForUser retrieves all cached bookmarks for a user
func (s *Service) GetCachedBookmarksForUser(ctx context.Context, userID uint) ([]database.Bookmark, error) {
	// This is a simplified implementation - in production, you might want to use Redis SCAN
	// For now, we'll return an empty slice as this would require more complex Redis operations
	return []database.Bookmark{}, nil
}

// QueueOfflineChange adds a change to the offline queue
func (s *Service) QueueOfflineChange(ctx context.Context, change *OfflineChange) error {
	key := fmt.Sprintf("offline:queue:%d:%s", change.UserID, change.ID)

	changeJSON, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("failed to marshal change: %w", err)
	}

	// Queue for 7 days
	err = s.redisClient.Set(ctx, key, string(changeJSON), time.Hour*24*7)
	if err != nil {
		return fmt.Errorf("failed to queue change: %w", err)
	}

	return nil
}

// GetOfflineQueue retrieves all queued changes for a user
func (s *Service) GetOfflineQueue(ctx context.Context, userID uint) ([]*OfflineChange, error) {
	// This is a simplified implementation - in production, you'd use Redis SCAN
	// For testing purposes, we'll simulate getting changes
	changes := []*OfflineChange{}

	// In a real implementation, you would scan for keys matching "offline:queue:userID:*"
	// and retrieve all changes. For now, we'll return empty slice.

	return changes, nil
}

// ProcessOfflineQueue processes all queued changes when connectivity is restored
func (s *Service) ProcessOfflineQueue(ctx context.Context, userID uint) error {
	changes, err := s.GetOfflineQueue(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get offline queue: %w", err)
	}

	for _, change := range changes {
		if change.Applied {
			continue
		}

		err := s.applyChange(ctx, change)
		if err != nil {
			// Log error but continue processing other changes
			continue
		}

		// Mark change as applied
		change.Applied = true
		if err := s.updateQueuedChange(ctx, change); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}

// applyChange applies a single offline change
func (s *Service) applyChange(ctx context.Context, change *OfflineChange) error {
	switch change.Type {
	case "bookmark_create":
		return s.applyBookmarkCreate(ctx, change)
	case "bookmark_update":
		return s.applyBookmarkUpdate(ctx, change)
	case "bookmark_delete":
		return s.applyBookmarkDelete(ctx, change)
	default:
		return ErrInvalidChangeType
	}
}

// applyBookmarkCreate applies a bookmark creation change
func (s *Service) applyBookmarkCreate(ctx context.Context, change *OfflineChange) error {
	var bookmarkData map[string]interface{}
	if err := json.Unmarshal([]byte(change.Data), &bookmarkData); err != nil {
		return fmt.Errorf("failed to unmarshal bookmark data: %w", err)
	}

	bookmark := database.Bookmark{
		UserID:      change.UserID,
		URL:         bookmarkData["url"].(string),
		Title:       bookmarkData["title"].(string),
		Description: bookmarkData["description"].(string),
		Tags:        bookmarkData["tags"].(string),
	}

	if err := s.db.Create(&bookmark).Error; err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	return nil
}

// applyBookmarkUpdate applies a bookmark update change
func (s *Service) applyBookmarkUpdate(ctx context.Context, change *OfflineChange) error {
	var bookmarkData map[string]interface{}
	if err := json.Unmarshal([]byte(change.Data), &bookmarkData); err != nil {
		return fmt.Errorf("failed to unmarshal bookmark data: %w", err)
	}

	// Update the bookmark in database
	err := s.db.Model(&database.Bookmark{}).
		Where("id = ? AND user_id = ?", change.ResourceID, change.UserID).
		Updates(bookmarkData).Error

	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	return nil
}

// applyBookmarkDelete applies a bookmark deletion change
func (s *Service) applyBookmarkDelete(ctx context.Context, change *OfflineChange) error {
	err := s.db.Where("id = ? AND user_id = ?", change.ResourceID, change.UserID).
		Delete(&database.Bookmark{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	return nil
}

// updateQueuedChange updates a change in the queue
func (s *Service) updateQueuedChange(ctx context.Context, change *OfflineChange) error {
	key := fmt.Sprintf("offline:queue:%d:%s", change.UserID, change.ID)

	changeJSON, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("failed to marshal change: %w", err)
	}

	err = s.redisClient.Set(ctx, key, string(changeJSON), time.Hour*24*7)
	if err != nil {
		return fmt.Errorf("failed to update queued change: %w", err)
	}

	return nil
}

// CheckConnectivity checks if the service is online
func (s *Service) CheckConnectivity(ctx context.Context) bool {
	// Simple connectivity check - try to make a HEAD request to a reliable endpoint
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head("https://www.google.com")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetOfflineStatus gets the current offline status for a user
func (s *Service) GetOfflineStatus(ctx context.Context, userID uint) (string, error) {
	key := fmt.Sprintf("offline:status:%d", userID)

	status, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" || err.Error() == "key not found" {
			return "online", nil // Default to online
		}
		return "", fmt.Errorf("failed to get offline status: %w", err)
	}

	return status, nil
}

// SetOfflineStatus sets the offline status for a user
func (s *Service) SetOfflineStatus(ctx context.Context, userID uint, status string) error {
	key := fmt.Sprintf("offline:status:%d", userID)

	err := s.redisClient.Set(ctx, key, status, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to set offline status: %w", err)
	}

	return nil
}

// CleanupExpiredCache removes expired cache entries
func (s *Service) CleanupExpiredCache(ctx context.Context, userID uint) error {
	// This is a simplified implementation
	// In production, you would use Redis SCAN to find and delete expired keys

	// For now, we'll just simulate cleanup
	return nil
}

// GetCacheStats returns cache statistics for a user
func (s *Service) GetCacheStats(ctx context.Context, userID uint) (*CacheStats, error) {
	key := fmt.Sprintf("offline:stats:%d", userID)

	statsJSON, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" || err.Error() == "key not found" {
			// Return default stats
			return &CacheStats{
				CachedBookmarksCount: 0,
				QueuedChangesCount:   0,
				LastSync:             time.Now(),
				CacheSize:            0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	var stats CacheStats
	err = json.Unmarshal([]byte(statsJSON), &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache stats: %w", err)
	}

	return &stats, nil
}

// UpdateCacheStats updates cache statistics for a user
func (s *Service) UpdateCacheStats(ctx context.Context, userID uint, stats *CacheStats) error {
	key := fmt.Sprintf("offline:stats:%d", userID)

	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal cache stats: %w", err)
	}

	err = s.redisClient.Set(ctx, key, string(statsJSON), time.Hour*24)
	if err != nil {
		return fmt.Errorf("failed to update cache stats: %w", err)
	}

	return nil
}

// ResolveConflict resolves conflicts between local and server changes
func (s *Service) ResolveConflict(localChange, serverChange *OfflineChange) *OfflineChange {
	// Simple conflict resolution: latest timestamp wins
	if serverChange.Timestamp.After(localChange.Timestamp) {
		return serverChange
	}
	return localChange
}

// SyncWhenOnline performs sync when connectivity is restored
func (s *Service) SyncWhenOnline(ctx context.Context, userID uint) error {
	// Check if we're online
	if !s.CheckConnectivity(ctx) {
		return errors.New("still offline")
	}

	// Update status to online
	if err := s.SetOfflineStatus(ctx, userID, "online"); err != nil {
		return fmt.Errorf("failed to set online status: %w", err)
	}

	// Process offline queue
	if err := s.ProcessOfflineQueue(ctx, userID); err != nil {
		return fmt.Errorf("failed to process offline queue: %w", err)
	}

	// Update cache stats
	stats, err := s.GetCacheStats(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	stats.LastSync = time.Now()
	if err := s.UpdateCacheStats(ctx, userID, stats); err != nil {
		return fmt.Errorf("failed to update cache stats: %w", err)
	}

	return nil
}

// CreateOfflineChange creates a new offline change
func (s *Service) CreateOfflineChange(userID uint, deviceID, changeType, resourceID, data string) *OfflineChange {
	return &OfflineChange{
		ID:         fmt.Sprintf("%d-%s-%d", userID, deviceID, time.Now().UnixNano()),
		UserID:     userID,
		DeviceID:   deviceID,
		Type:       changeType,
		ResourceID: resourceID,
		Data:       data,
		Timestamp:  time.Now(),
		Applied:    false,
	}
}

// ValidateChangeType validates if a change type is supported
func (s *Service) ValidateChangeType(changeType string) bool {
	validTypes := []string{
		"bookmark_create",
		"bookmark_update",
		"bookmark_delete",
		"collection_create",
		"collection_update",
		"collection_delete",
	}

	for _, validType := range validTypes {
		if changeType == validType {
			return true
		}
	}

	return false
}

// GetOfflineIndicator returns offline indicator information
func (s *Service) GetOfflineIndicator(ctx context.Context, userID uint) (map[string]interface{}, error) {
	status, err := s.GetOfflineStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offline status: %w", err)
	}

	stats, err := s.GetCacheStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	isOnline := s.CheckConnectivity(ctx)

	indicator := map[string]interface{}{
		"status":             status,
		"is_online":          isOnline,
		"cached_bookmarks":   stats.CachedBookmarksCount,
		"queued_changes":     stats.QueuedChangesCount,
		"last_sync":          stats.LastSync,
		"connectivity_check": isOnline,
	}

	return indicator, nil
}
