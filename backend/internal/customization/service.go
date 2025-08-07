package customization

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bookmark-sync-service/backend/pkg/worker"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Database interface for dependency injection
type Database interface {
	Create(value any) *gorm.DB
	Find(dest any, conds ...any) *gorm.DB
	Where(query any, args ...any) Database
	First(dest any, conds ...any) *gorm.DB
	Save(value any) *gorm.DB
	Delete(value any, conds ...any) *gorm.DB
	Order(value any) Database
	Limit(limit int) Database
	Offset(offset int) Database
	Preload(query string, args ...any) Database
}

// Redis interface for caching
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Service handles customization features
type Service struct {
	db         Database
	redis      RedisClient
	workerPool *worker.WorkerPool
	logger     *zap.Logger
}

// NewService creates a new customization service
func NewService(db Database, redis RedisClient, workerPool *worker.WorkerPool, logger *zap.Logger) *Service {
	return &Service{
		db:         db,
		redis:      redis,
		workerPool: workerPool,
		logger:     logger,
	}
}

// CreateTheme creates a new theme
func (s *Service) CreateTheme(ctx context.Context, userID string, req *CreateThemeRequest) (*ThemeResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Serialize config
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, ErrInvalidThemeConfig
	}

	theme := &Theme{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		CreatorID:   userID,
		IsPublic:    req.IsPublic,
		Config:      string(configJSON),
		PreviewURL:  req.PreviewURL,
	}

	if err := theme.Validate(); err != nil {
		return nil, err
	}

	if err := s.db.Create(theme).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, ErrThemeAlreadyExists
		}
		return nil, fmt.Errorf("failed to create theme: %w", err)
	}

	return s.themeToResponse(theme), nil
}

// GetTheme retrieves a theme by ID
func (s *Service) GetTheme(ctx context.Context, userID string, themeID uint) (*ThemeResponse, error) {
	var theme Theme
	err := s.db.First(&theme, themeID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrThemeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get theme: %w", err)
	}

	// Check if user can access this theme
	if !theme.IsPublic && theme.CreatorID != userID {
		return nil, ErrUnauthorizedTheme
	}

	return s.themeToResponse(&theme), nil
}

// UpdateTheme updates an existing theme
func (s *Service) UpdateTheme(ctx context.Context, userID string, themeID uint, req *UpdateThemeRequest) (*ThemeResponse, error) {
	var theme Theme
	err := s.db.First(&theme, themeID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrThemeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get theme: %w", err)
	}

	// Check if user owns this theme
	if theme.CreatorID != userID {
		return nil, ErrUnauthorizedTheme
	}

	// Update fields
	if req.DisplayName != "" {
		theme.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		theme.Description = req.Description
	}
	if req.IsPublic != nil {
		theme.IsPublic = *req.IsPublic
	}
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, ErrInvalidThemeConfig
		}
		theme.Config = string(configJSON)
	}
	if req.PreviewURL != "" {
		theme.PreviewURL = req.PreviewURL
	}

	if err := theme.Validate(); err != nil {
		return nil, err
	}

	if err := s.db.Save(&theme).Error; err != nil {
		return nil, fmt.Errorf("failed to update theme: %w", err)
	}

	return s.themeToResponse(&theme), nil
}

// DeleteTheme deletes a theme
func (s *Service) DeleteTheme(ctx context.Context, userID string, themeID uint) error {
	var theme Theme
	err := s.db.First(&theme, themeID).Error
	if err == gorm.ErrRecordNotFound {
		return ErrThemeNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get theme: %w", err)
	}

	// Check if user owns this theme
	if theme.CreatorID != userID {
		return ErrUnauthorizedTheme
	}

	if err := s.db.Delete(&theme).Error; err != nil {
		return fmt.Errorf("failed to delete theme: %w", err)
	}

	return nil
}

// ListThemes lists themes with filtering and pagination
func (s *Service) ListThemes(ctx context.Context, req *ThemeListRequest) ([]ThemeResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	query := s.db

	// Apply filters
	if req.PublicOnly {
		query = query.Where("is_public = ?", true)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	// Count total
	var total int64
	if err := query.Find(&[]Theme{}).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count themes: %w", err)
	}

	// Apply sorting and pagination
	orderBy := fmt.Sprintf("%s %s", req.SortBy, req.SortOrder)
	query = query.Order(orderBy).Limit(req.Limit).Offset((req.Page - 1) * req.Limit)

	var themes []Theme
	if err := query.Find(&themes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list themes: %w", err)
	}

	responses := make([]ThemeResponse, len(themes))
	for i, theme := range themes {
		responses[i] = *s.themeToResponse(&theme)
	}

	return responses, total, nil
}

// GetUserPreferences retrieves user preferences
func (s *Service) GetUserPreferences(ctx context.Context, userID string) (*UserPreferencesResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user_preferences:%s", userID)
	if cached, err := s.redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var prefs UserPreferencesResponse
		if json.Unmarshal([]byte(cached), &prefs) == nil {
			return &prefs, nil
		}
	}

	var prefs UserPreferences
	err := s.db.First(&prefs, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		// Create default preferences
		prefs = UserPreferences{
			UserID:               userID,
			Language:             "en",
			Timezone:             "UTC",
			DateFormat:           "YYYY-MM-DD",
			TimeFormat:           "24h",
			GridSize:             "medium",
			ViewMode:             "grid",
			SortBy:               "created_at",
			SortOrder:            "desc",
			ShowThumbnails:       true,
			ShowDescriptions:     true,
			ShowTags:             true,
			AutoSync:             true,
			SyncInterval:         300,
			NotificationsEnabled: true,
			SoundEnabled:         false,
			CompactMode:          false,
			ShowSidebar:          true,
			SidebarWidth:         250,
		}

		if err := s.db.Create(&prefs).Error; err != nil {
			return nil, fmt.Errorf("failed to create default preferences: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	response := s.preferencesToResponse(&prefs)

	// Cache the result
	if cacheData, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, string(cacheData), 30*time.Minute)
	}

	return response, nil
}

// UpdateUserPreferences updates user preferences
func (s *Service) UpdateUserPreferences(ctx context.Context, userID string, req *UpdateUserPreferencesRequest) (*UserPreferencesResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	var prefs UserPreferences
	err := s.db.First(&prefs, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		// Create new preferences
		prefs = UserPreferences{UserID: userID}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Update fields
	if req.Language != "" {
		prefs.Language = req.Language
	}
	if req.Timezone != "" {
		prefs.Timezone = req.Timezone
	}
	if req.DateFormat != "" {
		prefs.DateFormat = req.DateFormat
	}
	if req.TimeFormat != "" {
		prefs.TimeFormat = req.TimeFormat
	}
	if req.GridSize != "" {
		prefs.GridSize = req.GridSize
	}
	if req.ViewMode != "" {
		prefs.ViewMode = req.ViewMode
	}
	if req.SortBy != "" {
		prefs.SortBy = req.SortBy
	}
	if req.SortOrder != "" {
		prefs.SortOrder = req.SortOrder
	}
	if req.ShowThumbnails != nil {
		prefs.ShowThumbnails = *req.ShowThumbnails
	}
	if req.ShowDescriptions != nil {
		prefs.ShowDescriptions = *req.ShowDescriptions
	}
	if req.ShowTags != nil {
		prefs.ShowTags = *req.ShowTags
	}
	if req.AutoSync != nil {
		prefs.AutoSync = *req.AutoSync
	}
	if req.SyncInterval != nil {
		prefs.SyncInterval = *req.SyncInterval
	}
	if req.NotificationsEnabled != nil {
		prefs.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.SoundEnabled != nil {
		prefs.SoundEnabled = *req.SoundEnabled
	}
	if req.CompactMode != nil {
		prefs.CompactMode = *req.CompactMode
	}
	if req.ShowSidebar != nil {
		prefs.ShowSidebar = *req.ShowSidebar
	}
	if req.SidebarWidth != nil {
		prefs.SidebarWidth = *req.SidebarWidth
	}
	if req.CustomCSS != "" {
		prefs.CustomCSS = req.CustomCSS
	}

	if err := prefs.Validate(); err != nil {
		return nil, err
	}

	if err := s.db.Save(&prefs).Error; err != nil {
		return nil, fmt.Errorf("failed to update user preferences: %w", err)
	}

	// Clear cache
	cacheKey := fmt.Sprintf("user_preferences:%s", userID)
	s.redis.Del(ctx, cacheKey)

	return s.preferencesToResponse(&prefs), nil
}

// GetUserTheme retrieves user's active theme
func (s *Service) GetUserTheme(ctx context.Context, userID string) (*UserThemeResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user_theme:%s", userID)
	if cached, err := s.redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var userTheme UserThemeResponse
		if json.Unmarshal([]byte(cached), &userTheme) == nil {
			return &userTheme, nil
		}
	}

	var userTheme UserTheme
	err := s.db.Preload("Theme").First(&userTheme, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrThemeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user theme: %w", err)
	}

	response := s.userThemeToResponse(&userTheme)

	// Cache the result
	if cacheData, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, string(cacheData), 30*time.Minute)
	}

	return response, nil
}

// SetUserTheme sets user's active theme
func (s *Service) SetUserTheme(ctx context.Context, userID string, req *SetUserThemeRequest) (*UserThemeResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Check if theme exists
	var theme Theme
	err := s.db.First(&theme, req.ThemeID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrThemeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get theme: %w", err)
	}

	// Check if theme is accessible
	if !theme.IsPublic && theme.CreatorID != userID {
		return nil, ErrThemeNotPublic
	}

	// Serialize config
	var configJSON string
	if req.Config != nil {
		configBytes, err := json.Marshal(req.Config)
		if err != nil {
			return nil, ErrInvalidThemeConfig
		}
		configJSON = string(configBytes)
	}

	// Check if user already has a theme
	var userTheme UserTheme
	err = s.db.First(&userTheme, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		// Create new user theme
		userTheme = UserTheme{
			UserID:   userID,
			ThemeID:  req.ThemeID,
			IsActive: true,
			Config:   configJSON,
		}
		if err := s.db.Create(&userTheme).Error; err != nil {
			return nil, fmt.Errorf("failed to create user theme: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user theme: %w", err)
	} else {
		// Update existing user theme
		userTheme.ThemeID = req.ThemeID
		userTheme.Config = configJSON
		userTheme.IsActive = true
		if err := s.db.Save(&userTheme).Error; err != nil {
			return nil, fmt.Errorf("failed to update user theme: %w", err)
		}
	}

	// Load theme data
	if err := s.db.Preload("Theme").First(&userTheme, userTheme.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load user theme: %w", err)
	}

	// Clear cache
	cacheKey := fmt.Sprintf("user_theme:%s", userID)
	s.redis.Del(ctx, cacheKey)

	return s.userThemeToResponse(&userTheme), nil
}

// RateTheme rates a theme
func (s *Service) RateTheme(ctx context.Context, userID string, themeID uint, req *RateThemeRequest) (*ThemeRating, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Check if theme exists
	var theme Theme
	err := s.db.First(&theme, themeID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrThemeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get theme: %w", err)
	}

	// Check if user already rated this theme
	var existingRating ThemeRating
	err = s.db.First(&existingRating, "user_id = ? AND theme_id = ?", userID, themeID).Error
	if err == nil {
		return nil, ErrAlreadyRated
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing rating: %w", err)
	}

	// Create rating
	rating := &ThemeRating{
		UserID:  userID,
		ThemeID: themeID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	if err := rating.Validate(); err != nil {
		return nil, err
	}

	if err := s.db.Create(rating).Error; err != nil {
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	// Update theme rating statistics
	s.updateThemeRatingStats(themeID)

	return rating, nil
}

// Helper methods
func (s *Service) themeToResponse(theme *Theme) *ThemeResponse {
	var config any
	if theme.Config != "" {
		json.Unmarshal([]byte(theme.Config), &config)
	}

	return &ThemeResponse{
		ID:          theme.ID,
		Name:        theme.Name,
		DisplayName: theme.DisplayName,
		Description: theme.Description,
		CreatorID:   theme.CreatorID,
		IsPublic:    theme.IsPublic,
		IsDefault:   theme.IsDefault,
		Config:      config,
		PreviewURL:  theme.PreviewURL,
		Downloads:   theme.Downloads,
		Rating:      theme.Rating,
		RatingCount: theme.RatingCount,
		CreatedAt:   theme.CreatedAt,
		UpdatedAt:   theme.UpdatedAt,
	}
}

func (s *Service) preferencesToResponse(prefs *UserPreferences) *UserPreferencesResponse {
	return &UserPreferencesResponse{
		ID:                   prefs.ID,
		UserID:               prefs.UserID,
		Language:             prefs.Language,
		Timezone:             prefs.Timezone,
		DateFormat:           prefs.DateFormat,
		TimeFormat:           prefs.TimeFormat,
		GridSize:             prefs.GridSize,
		ViewMode:             prefs.ViewMode,
		SortBy:               prefs.SortBy,
		SortOrder:            prefs.SortOrder,
		ShowThumbnails:       prefs.ShowThumbnails,
		ShowDescriptions:     prefs.ShowDescriptions,
		ShowTags:             prefs.ShowTags,
		AutoSync:             prefs.AutoSync,
		SyncInterval:         prefs.SyncInterval,
		NotificationsEnabled: prefs.NotificationsEnabled,
		SoundEnabled:         prefs.SoundEnabled,
		CompactMode:          prefs.CompactMode,
		ShowSidebar:          prefs.ShowSidebar,
		SidebarWidth:         prefs.SidebarWidth,
		CustomCSS:            prefs.CustomCSS,
		CreatedAt:            prefs.CreatedAt,
		UpdatedAt:            prefs.UpdatedAt,
	}
}

func (s *Service) userThemeToResponse(userTheme *UserTheme) *UserThemeResponse {
	var config any
	if userTheme.Config != "" {
		json.Unmarshal([]byte(userTheme.Config), &config)
	}

	return &UserThemeResponse{
		ID:        userTheme.ID,
		UserID:    userTheme.UserID,
		ThemeID:   userTheme.ThemeID,
		Theme:     *s.themeToResponse(&userTheme.Theme),
		IsActive:  userTheme.IsActive,
		Config:    config,
		CreatedAt: userTheme.CreatedAt,
		UpdatedAt: userTheme.UpdatedAt,
	}
}

func (s *Service) updateThemeRatingStats(themeID uint) {
	// Submit theme rating update job to worker queue
	if s.workerPool != nil {
		job := worker.NewThemeRatingUpdateJob(themeID, s, s.logger)
		if err := s.workerPool.Submit(job); err != nil {
			s.logger.Warn("Failed to submit theme rating update job", zap.Error(err))
		}
	}
}

// UpdateThemeRating updates theme rating statistics (called by worker)
func (s *Service) UpdateThemeRating(ctx context.Context, themeID uint) error {
	var ratings []ThemeRating
	if err := s.db.Find(&ratings, "theme_id = ?", themeID).Error; err != nil {
		return fmt.Errorf("failed to fetch theme ratings: %w", err)
	}

	if len(ratings) == 0 {
		return nil
	}

	var totalRating int
	for _, rating := range ratings {
		totalRating += rating.Rating
	}

	avgRating := float64(totalRating) / float64(len(ratings))

	var theme Theme
	if err := s.db.First(&theme, themeID).Error; err == nil {
		theme.Rating = avgRating
		theme.RatingCount = len(ratings)
		if err := s.db.Save(&theme).Error; err != nil {
			return fmt.Errorf("failed to update theme rating: %w", err)
		}
	}

	return nil
}
