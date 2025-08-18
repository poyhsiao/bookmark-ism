package automation

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// WebhookEvent represents different types of webhook events
type WebhookEvent string

const (
	WebhookEventBookmarkCreated   WebhookEvent = "bookmark.created"
	WebhookEventBookmarkUpdated   WebhookEvent = "bookmark.updated"
	WebhookEventBookmarkDeleted   WebhookEvent = "bookmark.deleted"
	WebhookEventCollectionCreated WebhookEvent = "collection.created"
	WebhookEventCollectionUpdated WebhookEvent = "collection.updated"
	WebhookEventCollectionDeleted WebhookEvent = "collection.deleted"
	WebhookEventUserRegistered    WebhookEvent = "user.registered"
	WebhookEventUserUpdated       WebhookEvent = "user.updated"
)

// StringSlice is a custom type for handling JSON arrays in SQLite
type StringSlice []string

// Scan implements the Scanner interface for database deserialization
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	case []byte:
		return json.Unmarshal(v, s)
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}
}

// Value implements the Valuer interface for database serialization
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// StringMap is a custom type for handling JSON objects in SQLite
type StringMap map[string]string

// Scan implements the Scanner interface for database deserialization
func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = StringMap{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	default:
		return fmt.Errorf("cannot scan %T into StringMap", value)
	}
}

// Value implements the Valuer interface for database serialization
func (m StringMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	return json.Marshal(m)
}

// InterfaceMap is a custom type for handling JSON objects with mixed values in SQLite
type InterfaceMap map[string]interface{}

// Scan implements the Scanner interface for database deserialization
func (m *InterfaceMap) Scan(value interface{}) error {
	if value == nil {
		*m = InterfaceMap{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	default:
		return fmt.Errorf("cannot scan %T into InterfaceMap", value)
	}
}

// Value implements the Valuer interface for database serialization
func (m InterfaceMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	return json.Marshal(m)
}

// UintSlice is a custom type for handling JSON arrays of uints in SQLite
type UintSlice []uint

// Scan implements the Scanner interface for database deserialization
func (s *UintSlice) Scan(value interface{}) error {
	if value == nil {
		*s = UintSlice{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	case []byte:
		return json.Unmarshal(v, s)
	default:
		return fmt.Errorf("cannot scan %T into UintSlice", value)
	}
}

// Value implements the Valuer interface for database serialization
func (s UintSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Updated models with SQLite-compatible JSON fields

// WebhookEndpoint represents a webhook endpoint configuration
type WebhookEndpoint struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     string         `json:"user_id" gorm:"not null;index"`
	Name       string         `json:"name" gorm:"not null"`
	URL        string         `json:"url" gorm:"not null"`
	Secret     string         `json:"-" gorm:"not null"` // Hidden from JSON
	Events     StringSlice    `json:"events" gorm:"type:text"`
	Active     bool           `json:"active" gorm:"default:true"`
	RetryCount int            `json:"retry_count" gorm:"default:3"`
	Timeout    int            `json:"timeout" gorm:"default:30"` // seconds
	Headers    StringMap      `json:"headers" gorm:"type:text"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	EndpointID   uint           `json:"endpoint_id" gorm:"not null;index"`
	Event        WebhookEvent   `json:"event" gorm:"not null"`
	Payload      InterfaceMap   `json:"payload" gorm:"type:text"`
	Status       string         `json:"status" gorm:"not null"` // pending, success, failed
	StatusCode   int            `json:"status_code"`
	Response     string         `json:"response"`
	Error        string         `json:"error"`
	AttemptCount int            `json:"attempt_count" gorm:"default:0"`
	NextRetryAt  *time.Time     `json:"next_retry_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// RSSFeed represents an RSS feed configuration
type RSSFeed struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      string         `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	Link        string         `json:"link" gorm:"not null"`
	Language    string         `json:"language" gorm:"default:en"`
	Copyright   string         `json:"copyright"`
	Category    string         `json:"category"`
	TTL         int            `json:"ttl" gorm:"default:60"` // minutes
	MaxItems    int            `json:"max_items" gorm:"default:50"`
	Active      bool           `json:"active" gorm:"default:true"`
	PublicKey   string         `json:"public_key" gorm:"unique;not null"`
	Collections UintSlice      `json:"collections" gorm:"type:text"` // Collection IDs
	Tags        StringSlice    `json:"tags" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// BulkOperation represents a bulk operation job
type BulkOperation struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         string         `json:"user_id" gorm:"not null;index"`
	Type           string         `json:"type" gorm:"not null"`      // import, export, delete, update
	Status         string         `json:"status" gorm:"not null"`    // pending, running, completed, failed
	Progress       int            `json:"progress" gorm:"default:0"` // 0-100
	TotalItems     int            `json:"total_items" gorm:"default:0"`
	ProcessedItems int            `json:"processed_items" gorm:"default:0"`
	FailedItems    int            `json:"failed_items" gorm:"default:0"`
	Parameters     InterfaceMap   `json:"parameters" gorm:"type:text"`
	Result         InterfaceMap   `json:"result" gorm:"type:text"`
	Error          string         `json:"error"`
	StartedAt      *time.Time     `json:"started_at"`
	CompletedAt    *time.Time     `json:"completed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// BackupJob represents a backup job
type BackupJob struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        string         `json:"user_id" gorm:"not null;index"`
	Type          string         `json:"type" gorm:"not null"`   // full, incremental
	Status        string         `json:"status" gorm:"not null"` // pending, running, completed, failed
	Size          int64          `json:"size" gorm:"default:0"`  // bytes
	FilePath      string         `json:"file_path"`
	Checksum      string         `json:"checksum"`
	Compression   string         `json:"compression" gorm:"default:gzip"`
	Encrypted     bool           `json:"encrypted" gorm:"default:false"`
	RetentionDays int            `json:"retention_days" gorm:"default:30"`
	Error         string         `json:"error"`
	StartedAt     *time.Time     `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// APIIntegration represents an external API integration
type APIIntegration struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       string         `json:"user_id" gorm:"not null;index"`
	Name         string         `json:"name" gorm:"not null"`
	Type         string         `json:"type" gorm:"not null"` // pocket, instapaper, raindrop, etc.
	BaseURL      string         `json:"base_url" gorm:"not null"`
	APIKey       string         `json:"-" gorm:"not null"` // Hidden from JSON
	APISecret    string         `json:"-"`                 // Hidden from JSON
	AccessToken  string         `json:"-"`                 // Hidden from JSON
	RefreshToken string         `json:"-"`                 // Hidden from JSON
	TokenExpiry  *time.Time     `json:"-"`
	Active       bool           `json:"active" gorm:"default:true"`
	RateLimit    int            `json:"rate_limit" gorm:"default:100"` // requests per hour
	LastSync     *time.Time     `json:"last_sync"`
	SyncEnabled  bool           `json:"sync_enabled" gorm:"default:false"`
	SyncInterval int            `json:"sync_interval" gorm:"default:3600"` // seconds
	Config       InterfaceMap   `json:"config" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// AutomationRule represents an automation rule
type AutomationRule struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         string         `json:"user_id" gorm:"not null;index"`
	Name           string         `json:"name" gorm:"not null"`
	Description    string         `json:"description"`
	Trigger        string         `json:"trigger" gorm:"not null"` // bookmark_added, tag_added, etc.
	Conditions     InterfaceMap   `json:"conditions" gorm:"type:text"`
	Actions        InterfaceMap   `json:"actions" gorm:"type:text"`
	Active         bool           `json:"active" gorm:"default:true"`
	Priority       int            `json:"priority" gorm:"default:0"`
	ExecutionCount int            `json:"execution_count" gorm:"default:0"`
	LastExecuted   *time.Time     `json:"last_executed"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// WebhookPayload represents the structure of webhook payloads
type WebhookPayload struct {
	Event     WebhookEvent           `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// RSSItem represents an item in an RSS feed
type RSSItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Author      string    `json:"author,omitempty"`
	Category    string    `json:"category,omitempty"`
	GUID        string    `json:"guid"`
	PubDate     time.Time `json:"pub_date"`
	Tags        []string  `json:"tags,omitempty"`
}

// BulkOperationRequest represents a request for bulk operations
type BulkOperationRequest struct {
	Type       string                 `json:"type" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

// BackupRequest represents a backup request
type BackupRequest struct {
	Type        string `json:"type" binding:"required"` // full, incremental
	Compression string `json:"compression,omitempty"`   // gzip, zip, none
	Encrypted   bool   `json:"encrypted,omitempty"`
}

// APIIntegrationRequest represents an API integration request
type APIIntegrationRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	BaseURL      string                 `json:"base_url" binding:"required"`
	APIKey       string                 `json:"api_key" binding:"required"`
	APISecret    string                 `json:"api_secret,omitempty"`
	SyncEnabled  bool                   `json:"sync_enabled,omitempty"`
	SyncInterval int                    `json:"sync_interval,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// AutomationRuleRequest represents an automation rule request
type AutomationRuleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description,omitempty"`
	Trigger     string                 `json:"trigger" binding:"required"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Actions     map[string]interface{} `json:"actions" binding:"required"`
	Priority    int                    `json:"priority,omitempty"`
}
