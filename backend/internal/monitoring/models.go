package monitoring

import (
	"time"

	"gorm.io/gorm"
)

// LinkStatus represents the status of a link check
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

// CreateLinkCheckRequest represents the request to create a link check
type CreateLinkCheckRequest struct {
	BookmarkID uint   `json:"bookmark_id" binding:"required"`
	URL        string `json:"url" binding:"required,url"`
}

// CreateMonitoringJobRequest represents the request to create a monitoring job
type CreateMonitoringJobRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Enabled     bool   `json:"enabled"`
	Frequency   string `json:"frequency" binding:"required"` // cron expression
}

// UpdateMonitoringJobRequest represents the request to update a monitoring job
type UpdateMonitoringJobRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Enabled     *bool  `json:"enabled"`
	Frequency   string `json:"frequency" binding:"omitempty"`
}

// LinkCheckResponse represents the response for link check operations
type LinkCheckResponse struct {
	LinkCheck *LinkCheck `json:"link_check"`
	Message   string     `json:"message"`
}

// MonitoringJobResponse represents the response for monitoring job operations
type MonitoringJobResponse struct {
	Job     *LinkMonitoringJob `json:"job"`
	Message string             `json:"message"`
}

// MaintenanceReportResponse represents the response for maintenance report operations
type MaintenanceReportResponse struct {
	Report  *LinkMaintenanceReport `json:"report"`
	Message string                 `json:"message"`
}

// NotificationResponse represents the response for notification operations
type NotificationResponse struct {
	Notification *LinkChangeNotification `json:"notification"`
	Message      string                  `json:"message"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
