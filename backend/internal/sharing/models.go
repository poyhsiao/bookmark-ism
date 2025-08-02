package sharing

import (
	"time"

	"gorm.io/gorm"
)

// ShareType represents the type of sharing
type ShareType string

const (
	ShareTypePublic      ShareType = "public"
	ShareTypePrivate     ShareType = "private"
	ShareTypeShared      ShareType = "shared"
	ShareTypeCollaborate ShareType = "collaborate"
)

// SharePermission represents the permission level for sharing
type SharePermission string

const (
	PermissionView    SharePermission = "view"
	PermissionComment SharePermission = "comment"
	PermissionEdit    SharePermission = "edit"
	PermissionAdmin   SharePermission = "admin"
)

// CollectionShare represents a shared collection
type CollectionShare struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	CollectionID uint            `json:"collection_id" gorm:"not null;index"`
	UserID       uint            `json:"user_id" gorm:"not null;index"`
	ShareType    ShareType       `json:"share_type" gorm:"not null;default:'private'"`
	Permission   SharePermission `json:"permission" gorm:"not null;default:'view'"`
	ShareToken   string          `json:"share_token" gorm:"unique;not null;index"`
	Title        string          `json:"title" gorm:"size:255"`
	Description  string          `json:"description" gorm:"type:text"`
	Password     string          `json:"-" gorm:"size:255"` // Optional password protection
	ExpiresAt    *time.Time      `json:"expires_at"`
	ViewCount    int64           `json:"view_count" gorm:"default:0"`
	IsActive     bool            `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`
}

// CollectionCollaborator represents a collaborator on a shared collection
type CollectionCollaborator struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	CollectionID uint            `json:"collection_id" gorm:"not null;index"`
	UserID       uint            `json:"user_id" gorm:"not null;index"`
	InviterID    uint            `json:"inviter_id" gorm:"not null"`
	Permission   SharePermission `json:"permission" gorm:"not null;default:'view'"`
	Status       string          `json:"status" gorm:"not null;default:'pending'"` // pending, accepted, declined
	InvitedAt    time.Time       `json:"invited_at"`
	AcceptedAt   *time.Time      `json:"accepted_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`
}

// CollectionFork represents a forked collection
type CollectionFork struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	OriginalID        uint           `json:"original_id" gorm:"not null;index"`
	ForkedID          uint           `json:"forked_id" gorm:"not null;index"`
	UserID            uint           `json:"user_id" gorm:"not null;index"`
	ForkReason        string         `json:"fork_reason" gorm:"size:500"`
	PreserveBookmarks bool           `json:"preserve_bookmarks" gorm:"default:true"`
	PreserveStructure bool           `json:"preserve_structure" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// ShareActivity represents activity on shared collections
type ShareActivity struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ShareID      uint           `json:"share_id" gorm:"not null;index"`
	UserID       *uint          `json:"user_id" gorm:"index"`          // Nullable for anonymous views
	ActivityType string         `json:"activity_type" gorm:"not null"` // view, comment, edit, fork
	IPAddress    string         `json:"ip_address" gorm:"size:45"`
	UserAgent    string         `json:"user_agent" gorm:"size:500"`
	Metadata     string         `json:"metadata" gorm:"type:json"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateShareRequest represents a request to create a share
type CreateShareRequest struct {
	CollectionID uint            `json:"collection_id" binding:"required"`
	ShareType    ShareType       `json:"share_type" binding:"required,oneof=public private shared collaborate"`
	Permission   SharePermission `json:"permission" binding:"required,oneof=view comment edit admin"`
	Title        string          `json:"title" binding:"max=255"`
	Description  string          `json:"description" binding:"max=1000"`
	Password     string          `json:"password" binding:"max=255"`
	ExpiresAt    *time.Time      `json:"expires_at"`
}

// UpdateShareRequest represents a request to update a share
type UpdateShareRequest struct {
	ShareType   *ShareType       `json:"share_type,omitempty" binding:"omitempty,oneof=public private shared collaborate"`
	Permission  *SharePermission `json:"permission,omitempty" binding:"omitempty,oneof=view comment edit admin"`
	Title       *string          `json:"title,omitempty" binding:"omitempty,max=255"`
	Description *string          `json:"description,omitempty" binding:"omitempty,max=1000"`
	Password    *string          `json:"password,omitempty" binding:"omitempty,max=255"`
	ExpiresAt   *time.Time       `json:"expires_at,omitempty"`
	IsActive    *bool            `json:"is_active,omitempty"`
}

// ShareResponse represents a share response
type ShareResponse struct {
	ID           uint            `json:"id"`
	CollectionID uint            `json:"collection_id"`
	ShareType    ShareType       `json:"share_type"`
	Permission   SharePermission `json:"permission"`
	ShareToken   string          `json:"share_token"`
	ShareURL     string          `json:"share_url"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	HasPassword  bool            `json:"has_password"`
	ExpiresAt    *time.Time      `json:"expires_at"`
	ViewCount    int64           `json:"view_count"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CollaboratorRequest represents a request to add a collaborator
type CollaboratorRequest struct {
	Email      string          `json:"email" binding:"required,email"`
	Permission SharePermission `json:"permission" binding:"required,oneof=view comment edit admin"`
	Message    string          `json:"message" binding:"max=500"`
}

// ForkRequest represents a request to fork a collection
type ForkRequest struct {
	Name              string `json:"name" binding:"required,max=255"`
	Description       string `json:"description" binding:"max=1000"`
	ForkReason        string `json:"fork_reason" binding:"max=500"`
	PreserveBookmarks bool   `json:"preserve_bookmarks"`
	PreserveStructure bool   `json:"preserve_structure"`
}

// Validate validates the CreateShareRequest
func (r *CreateShareRequest) Validate() error {
	if r.CollectionID == 0 {
		return ErrInvalidCollectionID
	}

	if r.ShareType == "" {
		return ErrInvalidShareType
	}

	if r.Permission == "" {
		return ErrInvalidPermission
	}

	return nil
}

// Validate validates the UpdateShareRequest
func (r *UpdateShareRequest) Validate() error {
	if r.ShareType != nil && *r.ShareType == "" {
		return ErrInvalidShareType
	}

	if r.Permission != nil && *r.Permission == "" {
		return ErrInvalidPermission
	}

	return nil
}

// Validate validates the CollaboratorRequest
func (r *CollaboratorRequest) Validate() error {
	if r.Email == "" {
		return ErrInvalidEmail
	}

	if r.Permission == "" {
		return ErrInvalidPermission
	}

	return nil
}

// Validate validates the ForkRequest
func (r *ForkRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalidName
	}

	return nil
}

// ToResponse converts CollectionShare to ShareResponse
func (cs *CollectionShare) ToResponse(baseURL string) *ShareResponse {
	return &ShareResponse{
		ID:           cs.ID,
		CollectionID: cs.CollectionID,
		ShareType:    cs.ShareType,
		Permission:   cs.Permission,
		ShareToken:   cs.ShareToken,
		ShareURL:     baseURL + "/shared/" + cs.ShareToken,
		Title:        cs.Title,
		Description:  cs.Description,
		HasPassword:  cs.Password != "",
		ExpiresAt:    cs.ExpiresAt,
		ViewCount:    cs.ViewCount,
		IsActive:     cs.IsActive,
		CreatedAt:    cs.CreatedAt,
		UpdatedAt:    cs.UpdatedAt,
	}
}
