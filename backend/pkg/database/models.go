package database

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Import monitoring models
type LinkStatus string

const (
	LinkStatusActive   LinkStatus = "active"
	LinkStatusBroken   LinkStatus = "broken"
	LinkStatusRedirect LinkStatus = "redirect"
	LinkStatusTimeout  LinkStatus = "timeout"
	LinkStatusUnknown  LinkStatus = "unknown"
)

// LinkCheck represents a link monitoring check result
type LinkCheck struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	BookmarkID   uint           `json:"bookmark_id" gorm:"not null;index"`
	URL          string         `json:"url" gorm:"not null"`
	Status       LinkStatus     `json:"status" gorm:"not null"`
	StatusCode   int            `json:"status_code"`
	ResponseTime int64          `json:"response_time"` // in milliseconds
	RedirectURL  string         `json:"redirect_url,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CheckedAt    time.Time      `json:"checked_at" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// LinkMonitoringJob represents a scheduled monitoring job
type LinkMonitoringJob struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled" gorm:"default:true"`
	Frequency   string         `json:"frequency" gorm:"not null"` // cron expression
	LastRunAt   *time.Time     `json:"last_run_at"`
	NextRunAt   *time.Time     `json:"next_run_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// LinkMaintenanceReport represents a collection health report
type LinkMaintenanceReport struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	CollectionID  *uint          `json:"collection_id,omitempty" gorm:"index"`
	ReportType    string         `json:"report_type" gorm:"not null"` // "broken_links", "redirects", "duplicates"
	TotalLinks    int            `json:"total_links"`
	BrokenLinks   int            `json:"broken_links"`
	RedirectLinks int            `json:"redirect_links"`
	ActiveLinks   int            `json:"active_links"`
	Suggestions   string         `json:"suggestions" gorm:"type:text"` // Store as JSON string
	GeneratedAt   time.Time      `json:"generated_at" gorm:"not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// LinkChangeNotification represents a notification for link changes
type LinkChangeNotification struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`
	BookmarkID uint           `json:"bookmark_id" gorm:"not null;index"`
	ChangeType string         `json:"change_type" gorm:"not null"` // "broken", "redirect", "content_change"
	OldValue   string         `json:"old_value"`
	NewValue   string         `json:"new_value"`
	Message    string         `json:"message" gorm:"not null"`
	Read       bool           `json:"read" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// BaseModel contains common fields for all models
// 包含所有模型的通用字段，提供基本的 CRUD 時間戳和軟刪除功能
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"` // 主鍵 ID
	CreatedAt time.Time      `json:"created_at"`           // 創建時間
	UpdatedAt time.Time      `json:"updated_at"`           // 更新時間
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`       // 軟刪除時間（不在 JSON 中顯示）
}

// User represents a user in the system
// 代表系統中的用戶，包含基本信息、偏好設置和與其他實體的關聯
type User struct {
	BaseModel
	Email       string `gorm:"uniqueIndex;not null" json:"email"`    // 用戶電子郵件（唯一）
	Username    string `gorm:"uniqueIndex;not null" json:"username"` // 用戶名（唯一）
	DisplayName string `json:"display_name"`                         // 顯示名稱
	Avatar      string `json:"avatar,omitempty"`                     // 頭像 URL

	// Supabase Auth integration
	// Supabase 認證整合
	SupabaseID string `gorm:"uniqueIndex;not null" json:"supabase_id"` // Supabase 用戶 ID

	// User preferences (stored as JSON)
	// 用戶偏好設置（以 JSON 格式存儲）
	Preferences string `gorm:"type:jsonb" json:"preferences,omitempty"` // 用戶偏好設置

	// Relationships
	// 關聯關係
	Bookmarks   []Bookmark   `gorm:"foreignKey:UserID" json:"bookmarks,omitempty"`   // 用戶的書籤
	Collections []Collection `gorm:"foreignKey:UserID" json:"collections,omitempty"` // 用戶的收藏夾

	// Timestamps
	// 時間戳
	LastActiveAt *time.Time `json:"last_active_at,omitempty"` // 最後活躍時間
}

// Bookmark represents a bookmark in the system
type Bookmark struct {
	BaseModel
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	URL         string `gorm:"not null" json:"url"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description,omitempty"`
	Favicon     string `json:"favicon,omitempty"`
	Screenshot  string `json:"screenshot,omitempty"`

	// Metadata stored as JSON
	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Tags as JSON array
	Tags string `gorm:"type:jsonb" json:"tags,omitempty"`

	// Social metrics
	SaveCount    int `gorm:"default:0" json:"save_count"`
	LikeCount    int `gorm:"default:0" json:"like_count"`
	CommentCount int `gorm:"default:0" json:"comment_count"`

	// Status
	Status string `gorm:"default:'active'" json:"status"` // active, broken, redirected

	// Relationships
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Collections []Collection `gorm:"many2many:bookmark_collections;" json:"collections,omitempty"`
	Comments    []Comment    `gorm:"foreignKey:BookmarkID" json:"comments,omitempty"`

	// Timestamps
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	LastCheckedAt  *time.Time `json:"last_checked_at,omitempty"`
}

// Collection represents a bookmark collection
type Collection struct {
	BaseModel
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`

	// Hierarchy support
	ParentID *uint `gorm:"index" json:"parent_id,omitempty"`

	// Visibility and sharing
	Visibility string `gorm:"default:'private'" json:"visibility"` // private, public, shared
	ShareLink  string `gorm:"uniqueIndex" json:"share_link,omitempty"`

	// Metadata stored as JSON
	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Relationships
	User      User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Bookmarks []Bookmark   `gorm:"many2many:bookmark_collections;" json:"bookmarks,omitempty"`
	Children  []Collection `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Parent    *Collection  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

// Comment represents a comment on a bookmark
type Comment struct {
	BaseModel
	BookmarkID uint   `gorm:"not null;index" json:"bookmark_id"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Content    string `gorm:"not null" json:"content"`

	// Threading support
	ParentID *uint `gorm:"index" json:"parent_id,omitempty"`

	// Moderation
	IsModerated bool `gorm:"default:false" json:"is_moderated"`

	// Relationships
	Bookmark Bookmark  `gorm:"foreignKey:BookmarkID" json:"bookmark,omitempty"`
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Children []Comment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

// SyncEvent represents a synchronization event
type SyncEvent struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Type       string    `json:"type" gorm:"not null"`
	UserID     string    `json:"user_id" gorm:"not null;index"`
	ResourceID string    `json:"resource_id" gorm:"not null;index"`
	Action     string    `json:"action" gorm:"not null"`
	Data       string    `json:"data" gorm:"type:jsonb"`
	DeviceID   string    `json:"device_id" gorm:"not null;index"`
	Status     string    `json:"status" gorm:"default:'pending'"`
	Timestamp  time.Time `json:"timestamp" gorm:"not null;index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SyncState represents the synchronization state for a device
type SyncState struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"not null;index"`
	DeviceID     string    `json:"device_id" gorm:"not null;index"`
	LastSyncTime time.Time `json:"last_sync_time" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Follow represents a user following relationship
type Follow struct {
	BaseModel
	FollowerID  uint `gorm:"not null;index" json:"follower_id"`
	FollowingID uint `gorm:"not null;index" json:"following_id"`

	// Notification preferences stored as JSON
	NotificationSettings string `gorm:"type:jsonb" json:"notification_settings,omitempty"`

	// Relationships
	Follower  User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB) error {
	// Check if we're using PostgreSQL before enabling extensions
	// 檢查是否使用 PostgreSQL 再啟用擴展
	if db.Dialector.Name() == "postgres" {
		// Enable UUID extension
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
			return fmt.Errorf("failed to enable uuid-ossp extension: %w", err)
		}

		// Enable pgcrypto extension for encryption functions
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
			return fmt.Errorf("failed to enable pgcrypto extension: %w", err)
		}
	}

	// Run auto migrations
	if err := db.AutoMigrate(
		&User{},
		&Bookmark{},
		&Collection{},
		&Comment{},
		&SyncEvent{},
		&SyncState{},
		&Follow{},
		&CollectionShare{},
		&CollectionCollaborator{},
		&CollectionFork{},
		&ShareActivity{},
		&LinkCheck{},
		&LinkMonitoringJob{},
		&LinkMaintenanceReport{},
		&LinkChangeNotification{},
	); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	// Create additional indexes for performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// Rollback drops all tables (use with caution)
func Rollback(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&Follow{},
		&SyncState{},
		&SyncEvent{},
		&Comment{},
		&Collection{},
		&Bookmark{},
		&User{},
	)
}

// CollectionShare represents a shared collection
type CollectionShare struct {
	BaseModel
	CollectionID uint       `gorm:"not null;index" json:"collection_id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	ShareType    string     `gorm:"not null;default:'private'" json:"share_type"`
	Permission   string     `gorm:"not null;default:'view'" json:"permission"`
	ShareToken   string     `gorm:"unique;not null;index" json:"share_token"`
	Title        string     `gorm:"size:255" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	Password     string     `gorm:"size:255" json:"-"`
	ExpiresAt    *time.Time `json:"expires_at"`
	ViewCount    int64      `gorm:"default:0" json:"view_count"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
}

// CollectionCollaborator represents a collaborator on a shared collection
type CollectionCollaborator struct {
	BaseModel
	CollectionID uint       `gorm:"not null;index" json:"collection_id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	InviterID    uint       `gorm:"not null" json:"inviter_id"`
	Permission   string     `gorm:"not null;default:'view'" json:"permission"`
	Status       string     `gorm:"not null;default:'pending'" json:"status"`
	InvitedAt    time.Time  `json:"invited_at"`
	AcceptedAt   *time.Time `json:"accepted_at"`
}

// CollectionFork represents a forked collection
type CollectionFork struct {
	BaseModel
	OriginalID        uint   `gorm:"not null;index" json:"original_id"`
	ForkedID          uint   `gorm:"not null;index" json:"forked_id"`
	UserID            uint   `gorm:"not null;index" json:"user_id"`
	ForkReason        string `gorm:"size:500" json:"fork_reason"`
	PreserveBookmarks bool   `gorm:"default:true" json:"preserve_bookmarks"`
	PreserveStructure bool   `gorm:"default:true" json:"preserve_structure"`
}

// ShareActivity represents activity on shared collections
type ShareActivity struct {
	BaseModel
	ShareID      uint   `gorm:"not null;index" json:"share_id"`
	UserID       *uint  `gorm:"index" json:"user_id"`
	ActivityType string `gorm:"not null" json:"activity_type"`
	IPAddress    string `gorm:"size:45" json:"ip_address"`
	UserAgent    string `gorm:"size:500" json:"user_agent"`
	Metadata     string `gorm:"type:json" json:"metadata"`
}

// createIndexes creates additional database indexes for performance
func createIndexes(db *gorm.DB) error {
	// Basic indexes that work on both PostgreSQL and SQLite
	// 在 PostgreSQL 和 SQLite 上都能工作的基本索引
	basicIndexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_supabase_id ON users(supabase_id)",
		"CREATE INDEX IF NOT EXISTS idx_users_last_active ON users(last_active_at)",

		// Bookmark indexes
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_url ON bookmarks(url)",
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_status ON bookmarks(status)",
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_created_at ON bookmarks(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_user_created ON bookmarks(user_id, created_at)",

		// Collection indexes
		"CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_collections_parent_id ON collections(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_collections_visibility ON collections(visibility)",
		"CREATE INDEX IF NOT EXISTS idx_collections_share_link ON collections(share_link)",

		// Comment indexes
		"CREATE INDEX IF NOT EXISTS idx_comments_bookmark_id ON comments(bookmark_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id)",

		// Sync event indexes
		"CREATE INDEX IF NOT EXISTS idx_sync_events_user_device ON sync_events(user_id, device_id)",
		"CREATE INDEX IF NOT EXISTS idx_sync_events_status ON sync_events(status)",
		"CREATE INDEX IF NOT EXISTS idx_sync_events_timestamp ON sync_events(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_sync_events_created_at ON sync_events(created_at)",

		// Follow indexes
		"CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON follows(follower_id)",
		"CREATE INDEX IF NOT EXISTS idx_follows_following_id ON follows(following_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_follows_unique ON follows(follower_id, following_id)",
	}

	// PostgreSQL-specific indexes (full-text search)
	// PostgreSQL 特定索引（全文搜索）
	postgresIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_title_gin ON bookmarks USING gin(to_tsvector('english', title))",
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_description_gin ON bookmarks USING gin(to_tsvector('english', description))",
	}

	// Create basic indexes
	// 創建基本索引
	for _, index := range basicIndexes {
		if err := db.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", index, err)
		}
	}

	// Create PostgreSQL-specific indexes only if using PostgreSQL
	// 僅在使用 PostgreSQL 時創建 PostgreSQL 特定索引
	if db.Dialector.Name() == "postgres" {
		for _, index := range postgresIndexes {
			if err := db.Exec(index).Error; err != nil {
				return fmt.Errorf("failed to create index: %s, error: %w", index, err)
			}
		}
	}

	return nil
}

// SeedTestData seeds the database with test data for development
func SeedTestData(db *gorm.DB) error {
	// Check if data already exists
	var userCount int64
	if err := db.Model(&User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("database already contains data, skipping seed")
	}

	// Create test users
	testUsers := []User{
		{
			Email:       "admin@example.com",
			Username:    "admin",
			DisplayName: "Administrator",
			SupabaseID:  "test-admin-supabase-id",
			Preferences: `{"theme": "dark", "gridSize": "medium", "defaultView": "grid"}`,
		},
		{
			Email:       "user1@example.com",
			Username:    "user1",
			DisplayName: "Test User 1",
			SupabaseID:  "test-user1-supabase-id",
			Preferences: `{"theme": "light", "gridSize": "large", "defaultView": "list"}`,
		},
		{
			Email:       "user2@example.com",
			Username:    "user2",
			DisplayName: "Test User 2",
			SupabaseID:  "test-user2-supabase-id",
			Preferences: `{"theme": "auto", "gridSize": "small", "defaultView": "grid"}`,
		},
	}

	for _, user := range testUsers {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create test user %s: %w", user.Email, err)
		}
	}

	// Create test collections
	testCollections := []Collection{
		{
			UserID:      1,
			Name:        "Development Resources",
			Description: "Useful development tools and resources",
			Color:       "#3B82F6",
			Icon:        "code",
			Visibility:  "private",
			ShareLink:   "dev-resources-123",
		},
		{
			UserID:      1,
			Name:        "Design Inspiration",
			Description: "Beautiful designs and UI/UX resources",
			Color:       "#8B5CF6",
			Icon:        "palette",
			Visibility:  "public",
			ShareLink:   "design-inspiration-456",
		},
		{
			UserID:      2,
			Name:        "Learning Materials",
			Description: "Educational content and tutorials",
			Color:       "#10B981",
			Icon:        "book",
			Visibility:  "private",
			ShareLink:   "learning-materials-789",
		},
	}

	for _, collection := range testCollections {
		if err := db.Create(&collection).Error; err != nil {
			return fmt.Errorf("failed to create test collection %s: %w", collection.Name, err)
		}
	}

	// Create test bookmarks
	testBookmarks := []Bookmark{
		{
			UserID:      1,
			URL:         "https://github.com",
			Title:       "GitHub",
			Description: "The world's leading software development platform",
			Tags:        `["development", "git", "collaboration"]`,
			Status:      "active",
		},
		{
			UserID:      1,
			URL:         "https://stackoverflow.com",
			Title:       "Stack Overflow",
			Description: "The largest online community for developers",
			Tags:        `["development", "programming", "help"]`,
			Status:      "active",
		},
		{
			UserID:      2,
			URL:         "https://developer.mozilla.org",
			Title:       "MDN Web Docs",
			Description: "Resources for developers, by developers",
			Tags:        `["documentation", "web", "javascript"]`,
			Status:      "active",
		},
		{
			UserID:      2,
			URL:         "https://go.dev",
			Title:       "The Go Programming Language",
			Description: "Build fast, reliable, and efficient software at scale",
			Tags:        `["golang", "programming", "backend"]`,
			Status:      "active",
		},
	}

	for _, bookmark := range testBookmarks {
		if err := db.Create(&bookmark).Error; err != nil {
			return fmt.Errorf("failed to create test bookmark %s: %w", bookmark.Title, err)
		}
	}

	// Associate bookmarks with collections
	associations := []struct {
		BookmarkID   uint
		CollectionID uint
	}{
		{1, 1}, // GitHub -> Development Resources
		{2, 1}, // Stack Overflow -> Development Resources
		{3, 3}, // MDN -> Learning Materials
		{4, 3}, // Go -> Learning Materials
	}

	for _, assoc := range associations {
		var bookmark Bookmark
		var collection Collection

		if err := db.First(&bookmark, assoc.BookmarkID).Error; err != nil {
			return fmt.Errorf("failed to find bookmark %d: %w", assoc.BookmarkID, err)
		}

		if err := db.First(&collection, assoc.CollectionID).Error; err != nil {
			return fmt.Errorf("failed to find collection %d: %w", assoc.CollectionID, err)
		}

		if err := db.Model(&collection).Association("Bookmarks").Append(&bookmark); err != nil {
			return fmt.Errorf("failed to associate bookmark %d with collection %d: %w", assoc.BookmarkID, assoc.CollectionID, err)
		}
	}

	return nil
}

// SetupTestDB creates a test database for testing
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open test database: %w", err)
	}

	// Run migrations
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	return db, nil
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}
