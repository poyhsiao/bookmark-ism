package customization

import (
	"time"

	"gorm.io/gorm"
)

// Theme represents a UI theme configuration
type Theme struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Name        string         `json:"name" gorm:"not null;uniqueIndex"`
	DisplayName string         `json:"display_name" gorm:"not null"`
	Description string         `json:"description"`
	CreatorID   string         `json:"creator_id" gorm:"index"`
	IsPublic    bool           `json:"is_public" gorm:"default:false"`
	IsDefault   bool           `json:"is_default" gorm:"default:false"`
	Config      string         `json:"config" gorm:"type:text"` // JSON configuration
	PreviewURL  string         `json:"preview_url"`
	Downloads   int            `json:"downloads" gorm:"default:0"`
	Rating      float64        `json:"rating" gorm:"default:0"`
	RatingCount int            `json:"rating_count" gorm:"default:0"`
}

// UserTheme represents a user's theme preferences
type UserTheme struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserID    string         `json:"user_id" gorm:"not null;uniqueIndex"`
	ThemeID   uint           `json:"theme_id" gorm:"not null"`
	Theme     Theme          `json:"theme" gorm:"foreignKey:ThemeID"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	Config    string         `json:"config" gorm:"type:text"` // User-specific overrides
}

// UserPreferences represents user interface preferences
type UserPreferences struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
	UserID               string         `json:"user_id" gorm:"not null;uniqueIndex"`
	Language             string         `json:"language" gorm:"default:'en'"`
	Timezone             string         `json:"timezone" gorm:"default:'UTC'"`
	DateFormat           string         `json:"date_format" gorm:"default:'YYYY-MM-DD'"`
	TimeFormat           string         `json:"time_format" gorm:"default:'24h'"`
	GridSize             string         `json:"grid_size" gorm:"default:'medium'"` // small, medium, large
	ViewMode             string         `json:"view_mode" gorm:"default:'grid'"`   // grid, list, compact
	SortBy               string         `json:"sort_by" gorm:"default:'created_at'"`
	SortOrder            string         `json:"sort_order" gorm:"default:'desc'"`
	ShowThumbnails       bool           `json:"show_thumbnails" gorm:"default:true"`
	ShowDescriptions     bool           `json:"show_descriptions" gorm:"default:true"`
	ShowTags             bool           `json:"show_tags" gorm:"default:true"`
	AutoSync             bool           `json:"auto_sync" gorm:"default:true"`
	SyncInterval         int            `json:"sync_interval" gorm:"default:300"` // seconds
	NotificationsEnabled bool           `json:"notifications_enabled" gorm:"default:true"`
	SoundEnabled         bool           `json:"sound_enabled" gorm:"default:false"`
	CompactMode          bool           `json:"compact_mode" gorm:"default:false"`
	ShowSidebar          bool           `json:"show_sidebar" gorm:"default:true"`
	SidebarWidth         int            `json:"sidebar_width" gorm:"default:250"` // pixels
	CustomCSS            string         `json:"custom_css" gorm:"type:text"`
}

// ThemeRating represents user ratings for themes
type ThemeRating struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserID    string         `json:"user_id" gorm:"not null;index"`
	ThemeID   uint           `json:"theme_id" gorm:"not null;index"`
	Theme     Theme          `json:"theme" gorm:"foreignKey:ThemeID"`
	Rating    int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   string         `json:"comment"`
}

// Request/Response models
type CreateThemeRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsPublic    bool   `json:"is_public"`
	Config      any    `json:"config" binding:"required"`
	PreviewURL  string `json:"preview_url" binding:"omitempty,url"`
}

type UpdateThemeRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,min=3,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	IsPublic    *bool  `json:"is_public"`
	Config      any    `json:"config"`
	PreviewURL  string `json:"preview_url" binding:"omitempty,url"`
}

type UpdateUserPreferencesRequest struct {
	Language             string `json:"language" binding:"omitempty,oneof=en zh-CN zh-TW ja ko"`
	Timezone             string `json:"timezone"`
	DateFormat           string `json:"date_format" binding:"omitempty,oneof='YYYY-MM-DD' 'MM/DD/YYYY' 'DD/MM/YYYY'"`
	TimeFormat           string `json:"time_format" binding:"omitempty,oneof='12h' '24h'"`
	GridSize             string `json:"grid_size" binding:"omitempty,oneof=small medium large"`
	ViewMode             string `json:"view_mode" binding:"omitempty,oneof=grid list compact"`
	SortBy               string `json:"sort_by" binding:"omitempty,oneof=created_at updated_at title url"`
	SortOrder            string `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	ShowThumbnails       *bool  `json:"show_thumbnails"`
	ShowDescriptions     *bool  `json:"show_descriptions"`
	ShowTags             *bool  `json:"show_tags"`
	AutoSync             *bool  `json:"auto_sync"`
	SyncInterval         *int   `json:"sync_interval" binding:"omitempty,min=60,max=3600"`
	NotificationsEnabled *bool  `json:"notifications_enabled"`
	SoundEnabled         *bool  `json:"sound_enabled"`
	CompactMode          *bool  `json:"compact_mode"`
	ShowSidebar          *bool  `json:"show_sidebar"`
	SidebarWidth         *int   `json:"sidebar_width" binding:"omitempty,min=200,max=500"`
	CustomCSS            string `json:"custom_css"`
}

type SetUserThemeRequest struct {
	ThemeID uint `json:"theme_id" binding:"required"`
	Config  any  `json:"config"`
}

type RateThemeRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=500"`
}

type ThemeListRequest struct {
	Page       int    `json:"page" binding:"omitempty,min=1"`
	Limit      int    `json:"limit" binding:"omitempty,min=1,max=100"`
	Search     string `json:"search"`
	Category   string `json:"category"`
	SortBy     string `json:"sort_by" binding:"omitempty,oneof=name created_at downloads rating"`
	SortOrder  string `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	PublicOnly bool   `json:"public_only"`
}

// Response models
type ThemeResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	CreatorID   string    `json:"creator_id"`
	IsPublic    bool      `json:"is_public"`
	IsDefault   bool      `json:"is_default"`
	Config      any       `json:"config"`
	PreviewURL  string    `json:"preview_url"`
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"`
	RatingCount int       `json:"rating_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserThemeResponse struct {
	ID        uint          `json:"id"`
	UserID    string        `json:"user_id"`
	ThemeID   uint          `json:"theme_id"`
	Theme     ThemeResponse `json:"theme"`
	IsActive  bool          `json:"is_active"`
	Config    any           `json:"config"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type UserPreferencesResponse struct {
	ID                   uint      `json:"id"`
	UserID               string    `json:"user_id"`
	Language             string    `json:"language"`
	Timezone             string    `json:"timezone"`
	DateFormat           string    `json:"date_format"`
	TimeFormat           string    `json:"time_format"`
	GridSize             string    `json:"grid_size"`
	ViewMode             string    `json:"view_mode"`
	SortBy               string    `json:"sort_by"`
	SortOrder            string    `json:"sort_order"`
	ShowThumbnails       bool      `json:"show_thumbnails"`
	ShowDescriptions     bool      `json:"show_descriptions"`
	ShowTags             bool      `json:"show_tags"`
	AutoSync             bool      `json:"auto_sync"`
	SyncInterval         int       `json:"sync_interval"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
	SoundEnabled         bool      `json:"sound_enabled"`
	CompactMode          bool      `json:"compact_mode"`
	ShowSidebar          bool      `json:"show_sidebar"`
	SidebarWidth         int       `json:"sidebar_width"`
	CustomCSS            string    `json:"custom_css"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Validation methods
func (t *Theme) Validate() error {
	if t.Name == "" {
		return ErrInvalidThemeName
	}
	if t.DisplayName == "" {
		return ErrInvalidDisplayName
	}
	if len(t.Name) < 3 || len(t.Name) > 50 {
		return ErrInvalidThemeName
	}
	if len(t.DisplayName) < 3 || len(t.DisplayName) > 100 {
		return ErrInvalidDisplayName
	}
	if len(t.Description) > 500 {
		return ErrInvalidDescription
	}
	return nil
}

func (up *UserPreferences) Validate() error {
	if up.UserID == "" {
		return ErrInvalidUserID
	}

	validLanguages := map[string]bool{
		"en": true, "zh-CN": true, "zh-TW": true, "ja": true, "ko": true,
	}
	if !validLanguages[up.Language] {
		return ErrInvalidLanguage
	}

	validGridSizes := map[string]bool{
		"small": true, "medium": true, "large": true,
	}
	if !validGridSizes[up.GridSize] {
		return ErrInvalidGridSize
	}

	validViewModes := map[string]bool{
		"grid": true, "list": true, "compact": true,
	}
	if !validViewModes[up.ViewMode] {
		return ErrInvalidViewMode
	}

	validSortBy := map[string]bool{
		"created_at": true, "updated_at": true, "title": true, "url": true,
	}
	if !validSortBy[up.SortBy] {
		return ErrInvalidSortBy
	}

	validSortOrder := map[string]bool{
		"asc": true, "desc": true,
	}
	if !validSortOrder[up.SortOrder] {
		return ErrInvalidSortOrder
	}

	if up.SyncInterval < 60 || up.SyncInterval > 3600 {
		return ErrInvalidSyncInterval
	}

	if up.SidebarWidth < 200 || up.SidebarWidth > 500 {
		return ErrInvalidSidebarWidth
	}

	return nil
}

func (tr *ThemeRating) Validate() error {
	if tr.UserID == "" {
		return ErrInvalidUserID
	}
	if tr.ThemeID == 0 {
		return ErrInvalidThemeID
	}
	if tr.Rating < 1 || tr.Rating > 5 {
		return ErrInvalidRating
	}
	if len(tr.Comment) > 500 {
		return ErrInvalidComment
	}
	return nil
}
