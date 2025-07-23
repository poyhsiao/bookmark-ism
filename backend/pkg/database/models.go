package database

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents a user in the system
type User struct {
	BaseModel
	Email       string `gorm:"uniqueIndex;not null" json:"email"`
	Username    string `gorm:"uniqueIndex;not null" json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar,omitempty"`

	// Supabase Auth integration
	SupabaseID string `gorm:"uniqueIndex;not null" json:"supabase_id"`

	// User preferences (stored as JSON)
	Preferences string `gorm:"type:jsonb" json:"preferences,omitempty"`

	// Relationships
	Bookmarks   []Bookmark   `gorm:"foreignKey:UserID" json:"bookmarks,omitempty"`
	Collections []Collection `gorm:"foreignKey:UserID" json:"collections,omitempty"`

	// Timestamps
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
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
	BaseModel
	UserID       uint   `gorm:"not null;index" json:"user_id"`
	DeviceID     string `gorm:"not null;index" json:"device_id"`
	EventType    string `gorm:"not null" json:"event_type"`    // create, update, delete
	ResourceType string `gorm:"not null" json:"resource_type"` // bookmark, collection, tag
	ResourceID   uint   `gorm:"not null" json:"resource_id"`

	// Change data stored as JSON
	Changes string `gorm:"type:jsonb" json:"changes,omitempty"`

	// Processing status
	Processed        bool `gorm:"default:false" json:"processed"`
	ConflictResolved bool `gorm:"default:false" json:"conflict_resolved"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
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
	return db.AutoMigrate(
		&User{},
		&Bookmark{},
		&Collection{},
		&Comment{},
		&SyncEvent{},
		&Follow{},
	)
}
