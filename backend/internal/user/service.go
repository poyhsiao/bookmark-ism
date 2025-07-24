package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/storage"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service handles user profile operations
type Service struct {
	db            *gorm.DB
	storageClient *storage.Client
	logger        *zap.Logger
}

// NewService creates a new user service
func NewService(db *gorm.DB, storageClient *storage.Client, logger *zap.Logger) *Service {
	return &Service{
		db:            db,
		storageClient: storageClient,
		logger:        logger,
	}
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Theme       string `json:"theme"`       // light, dark, auto
	GridSize    string `json:"gridSize"`    // small, medium, large
	DefaultView string `json:"defaultView"` // grid, list
	Language    string `json:"language"`    // en, zh-CN, zh-TW
	Timezone    string `json:"timezone"`    // UTC offset or timezone name
}

// UserQuotas represents user quotas and limits
type UserQuotas struct {
	MaxBookmarks   int `json:"max_bookmarks"`
	MaxCollections int `json:"max_collections"`
	StorageLimit   int `json:"storage_limit"`  // in bytes
	APIRateLimit   int `json:"api_rate_limit"` // requests per minute
}

// UserProfile represents a user's profile information
type UserProfile struct {
	ID           uint            `json:"id"`
	Email        string          `json:"email"`
	Username     string          `json:"username"`
	DisplayName  string          `json:"display_name"`
	Avatar       string          `json:"avatar,omitempty"`
	Preferences  UserPreferences `json:"preferences"`
	Quotas       UserQuotas      `json:"quotas"`
	Stats        UserStats       `json:"stats"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	LastActiveAt *time.Time      `json:"last_active_at,omitempty"`
}

// UserStats represents user statistics
type UserStats struct {
	BookmarkCount   int `json:"bookmark_count"`
	CollectionCount int `json:"collection_count"`
	StorageUsed     int `json:"storage_used"` // in bytes
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty" binding:"omitempty,min=1,max=100"`
	Username    string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
}

// UpdatePreferencesRequest represents a preferences update request
type UpdatePreferencesRequest struct {
	Theme       string `json:"theme,omitempty" binding:"omitempty,oneof=light dark auto"`
	GridSize    string `json:"gridSize,omitempty" binding:"omitempty,oneof=small medium large"`
	DefaultView string `json:"defaultView,omitempty" binding:"omitempty,oneof=grid list"`
	Language    string `json:"language,omitempty" binding:"omitempty,oneof=en zh-CN zh-TW"`
	Timezone    string `json:"timezone,omitempty"`
}

// GetProfile retrieves a user's profile
func (s *Service) GetProfile(ctx context.Context, userID uint) (*UserProfile, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Parse preferences
	preferences := UserPreferences{
		Theme:       "light",
		GridSize:    "medium",
		DefaultView: "grid",
		Language:    "en",
		Timezone:    "UTC",
	}
	if user.Preferences != "" {
		if err := json.Unmarshal([]byte(user.Preferences), &preferences); err != nil {
			s.logger.Warn("Failed to parse user preferences", zap.Error(err), zap.Uint("user_id", userID))
		}
	}

	// Get user statistics
	stats, err := s.getUserStats(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user stats", zap.Error(err), zap.Uint("user_id", userID))
		stats = &UserStats{}
	}

	// Set default quotas (in a real implementation, these might be configurable)
	quotas := UserQuotas{
		MaxBookmarks:   10000,
		MaxCollections: 1000,
		StorageLimit:   1024 * 1024 * 1024, // 1GB
		APIRateLimit:   1000,               // 1000 requests per minute
	}

	return &UserProfile{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Avatar:       user.Avatar,
		Preferences:  preferences,
		Quotas:       quotas,
		Stats:        *stats,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		LastActiveAt: user.LastActiveAt,
	}, nil
}

// UpdateProfile updates a user's profile information
func (s *Service) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*UserProfile, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if username is already taken (if being updated)
	if req.Username != "" && req.Username != user.Username {
		var existingUser database.User
		if err := s.db.Where("username = ? AND id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("username already taken")
		}
	}

	// Update fields
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Username != "" {
		user.Username = req.Username
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("User profile updated", zap.Uint("user_id", userID))

	return s.GetProfile(ctx, userID)
}

// UpdatePreferences updates a user's preferences
func (s *Service) UpdatePreferences(ctx context.Context, userID uint, req *UpdatePreferencesRequest) (*UserProfile, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Parse existing preferences
	preferences := UserPreferences{
		Theme:       "light",
		GridSize:    "medium",
		DefaultView: "grid",
		Language:    "en",
		Timezone:    "UTC",
	}
	if user.Preferences != "" {
		if err := json.Unmarshal([]byte(user.Preferences), &preferences); err != nil {
			s.logger.Warn("Failed to parse existing preferences", zap.Error(err), zap.Uint("user_id", userID))
		}
	}

	// Update preferences
	if req.Theme != "" {
		preferences.Theme = req.Theme
	}
	if req.GridSize != "" {
		preferences.GridSize = req.GridSize
	}
	if req.DefaultView != "" {
		preferences.DefaultView = req.DefaultView
	}
	if req.Language != "" {
		preferences.Language = req.Language
	}
	if req.Timezone != "" {
		preferences.Timezone = req.Timezone
	}

	// Save preferences
	preferencesJSON, err := json.Marshal(preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preferences: %w", err)
	}

	user.Preferences = string(preferencesJSON)
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	s.logger.Info("User preferences updated", zap.Uint("user_id", userID))

	return s.GetProfile(ctx, userID)
}

// UploadAvatar uploads a user's avatar image
func (s *Service) UploadAvatar(ctx context.Context, userID uint, imageData []byte, contentType string) (*UserProfile, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Upload image to storage
	avatarKey := fmt.Sprintf("avatars/user_%d_%d", userID, time.Now().Unix())
	avatarURL, err := s.storageClient.UploadFile(ctx, avatarKey, imageData, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Update user avatar
	user.Avatar = avatarURL
	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	s.logger.Info("User avatar updated", zap.Uint("user_id", userID), zap.String("avatar_url", avatarURL))

	return s.GetProfile(ctx, userID)
}

// ExportUserData exports all user data for GDPR compliance
func (s *Service) ExportUserData(ctx context.Context, userID uint) (map[string]interface{}, error) {
	var user database.User
	if err := s.db.Preload("Bookmarks").Preload("Collections").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user statistics
	stats, err := s.getUserStats(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user stats for export", zap.Error(err), zap.Uint("user_id", userID))
		stats = &UserStats{}
	}

	exportData := map[string]interface{}{
		"user_profile": map[string]interface{}{
			"id":             user.ID,
			"email":          user.Email,
			"username":       user.Username,
			"display_name":   user.DisplayName,
			"avatar":         user.Avatar,
			"preferences":    user.Preferences,
			"created_at":     user.CreatedAt,
			"updated_at":     user.UpdatedAt,
			"last_active_at": user.LastActiveAt,
		},
		"bookmarks":   user.Bookmarks,
		"collections": user.Collections,
		"statistics":  stats,
		"export_date": time.Now().UTC(),
	}

	s.logger.Info("User data exported", zap.Uint("user_id", userID))

	return exportData, nil
}

// DeleteUser deletes a user account and all associated data
func (s *Service) DeleteUser(ctx context.Context, userID uint) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete user's bookmarks
	if err := tx.Where("user_id = ?", userID).Delete(&database.Bookmark{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user bookmarks: %w", err)
	}

	// Delete user's collections
	if err := tx.Where("user_id = ?", userID).Delete(&database.Collection{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user collections: %w", err)
	}

	// Delete user's comments
	if err := tx.Where("user_id = ?", userID).Delete(&database.Comment{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user comments: %w", err)
	}

	// Delete user's sync events
	if err := tx.Where("user_id = ?", userID).Delete(&database.SyncEvent{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user sync events: %w", err)
	}

	// Delete user's follows
	if err := tx.Where("follower_id = ? OR following_id = ?", userID, userID).Delete(&database.Follow{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user follows: %w", err)
	}

	// Delete user
	if err := tx.Delete(&database.User{}, userID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit user deletion: %w", err)
	}

	s.logger.Info("User account deleted", zap.Uint("user_id", userID))

	return nil
}

// getUserStats calculates user statistics
func (s *Service) getUserStats(ctx context.Context, userID uint) (*UserStats, error) {
	var stats UserStats

	// Count bookmarks
	var bookmarkCount int64
	if err := s.db.Model(&database.Bookmark{}).Where("user_id = ?", userID).Count(&bookmarkCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count bookmarks: %w", err)
	}
	stats.BookmarkCount = int(bookmarkCount)

	// Count collections
	var collectionCount int64
	if err := s.db.Model(&database.Collection{}).Where("user_id = ?", userID).Count(&collectionCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count collections: %w", err)
	}
	stats.CollectionCount = int(collectionCount)

	// For storage used, we would need to calculate the size of uploaded files
	// For now, we'll set it to 0
	stats.StorageUsed = 0

	return &stats, nil
}
