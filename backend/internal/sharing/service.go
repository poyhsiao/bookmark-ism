package sharing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
)

// Service represents the sharing service
type Service struct {
	db      *gorm.DB
	baseURL string
}

// NewService creates a new sharing service
func NewService(db *gorm.DB, baseURL string) *Service {
	return &Service{
		db:      db,
		baseURL: baseURL,
	}
}

// CreateShare creates a new collection share
func (s *Service) CreateShare(ctx context.Context, userID uint, request *CreateShareRequest) (*ShareResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Check if collection exists and user has access
	var collection database.Collection
	if err := s.db.First(&collection, request.CollectionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCollectionNotFound
		}
		return nil, fmt.Errorf("failed to find collection: %w", err)
	}

	// Check if user owns the collection or has admin permission
	if collection.UserID != userID {
		// TODO: Check if user has admin permission on this collection
		return nil, ErrUnauthorized
	}

	// Generate unique share token
	shareToken, err := s.generateShareToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate share token: %w", err)
	}

	// Create share
	share := &CollectionShare{
		CollectionID: request.CollectionID,
		UserID:       userID,
		ShareType:    request.ShareType,
		Permission:   request.Permission,
		ShareToken:   shareToken,
		Title:        request.Title,
		Description:  request.Description,
		Password:     request.Password, // TODO: Hash password
		ExpiresAt:    request.ExpiresAt,
		IsActive:     true,
	}

	if err := s.db.Create(share).Error; err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	return share.ToResponse(s.baseURL), nil
}

// GetShareByToken retrieves a share by its token
func (s *Service) GetShareByToken(ctx context.Context, token string) (*CollectionShare, error) {
	var share CollectionShare
	if err := s.db.First(&share, "share_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrShareNotFound
		}
		return nil, fmt.Errorf("failed to find share: %w", err)
	}

	// Check if share is active
	if !share.IsActive {
		return nil, ErrShareInactive
	}

	// Check if share has expired
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return nil, ErrShareExpired
	}

	return &share, nil
}

// UpdateShare updates an existing share
func (s *Service) UpdateShare(ctx context.Context, userID uint, shareID uint, request *UpdateShareRequest) (*ShareResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Get existing share
	var share CollectionShare
	if err := s.db.First(&share, shareID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrShareNotFound
		}
		return nil, fmt.Errorf("failed to find share: %w", err)
	}

	// Check if user owns the share
	if share.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields
	if request.ShareType != nil {
		share.ShareType = *request.ShareType
	}
	if request.Permission != nil {
		share.Permission = *request.Permission
	}
	if request.Title != nil {
		share.Title = *request.Title
	}
	if request.Description != nil {
		share.Description = *request.Description
	}
	if request.Password != nil {
		share.Password = *request.Password // TODO: Hash password
	}
	if request.ExpiresAt != nil {
		share.ExpiresAt = request.ExpiresAt
	}
	if request.IsActive != nil {
		share.IsActive = *request.IsActive
	}

	if err := s.db.Save(&share).Error; err != nil {
		return nil, fmt.Errorf("failed to update share: %w", err)
	}

	return share.ToResponse(s.baseURL), nil
}

// DeleteShare deletes a share
func (s *Service) DeleteShare(ctx context.Context, userID uint, shareID uint) error {
	// Get existing share
	var share CollectionShare
	if err := s.db.First(&share, shareID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrShareNotFound
		}
		return fmt.Errorf("failed to find share: %w", err)
	}

	// Check if user owns the share
	if share.UserID != userID {
		return ErrUnauthorized
	}

	if err := s.db.Delete(&share, shareID).Error; err != nil {
		return fmt.Errorf("failed to delete share: %w", err)
	}

	return nil
}

// GetUserShares retrieves all shares for a user
func (s *Service) GetUserShares(ctx context.Context, userID uint) ([]CollectionShare, error) {
	var shares []CollectionShare
	if err := s.db.Find(&shares, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user shares: %w", err)
	}

	return shares, nil
}

// GetCollectionShares retrieves all shares for a collection
func (s *Service) GetCollectionShares(ctx context.Context, userID uint, collectionID uint) ([]CollectionShare, error) {
	// Check if user owns the collection
	var collection database.Collection
	if err := s.db.First(&collection, collectionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCollectionNotFound
		}
		return nil, fmt.Errorf("failed to find collection: %w", err)
	}

	if collection.UserID != userID {
		return nil, ErrUnauthorized
	}

	var shares []CollectionShare
	if err := s.db.Find(&shares, "collection_id = ?", collectionID).Error; err != nil {
		return nil, fmt.Errorf("failed to get collection shares: %w", err)
	}

	return shares, nil
}

// ForkCollection creates a fork of a collection
func (s *Service) ForkCollection(ctx context.Context, userID uint, originalCollectionID uint, request *ForkRequest) (*database.Collection, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Get original collection
	var originalCollection database.Collection
	if err := s.db.Preload("Bookmarks").First(&originalCollection, originalCollectionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCollectionNotFound
		}
		return nil, fmt.Errorf("failed to find original collection: %w", err)
	}

	// Check if user is trying to fork their own collection
	if originalCollection.UserID == userID {
		return nil, ErrCannotForkOwnCollection
	}

	// TODO: Check if collection allows forking based on share settings

	var forkedCollection *database.Collection
	var fork *CollectionFork

	// Use transaction to ensure consistency
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create forked collection
		forkedCollection = &database.Collection{
			UserID:      userID,
			Name:        request.Name,
			Description: request.Description,
			Visibility:  "private", // Forked collections are private by default
		}

		if err := tx.Create(forkedCollection).Error; err != nil {
			return fmt.Errorf("failed to create forked collection: %w", err)
		}

		// Copy bookmarks if requested
		if request.PreserveBookmarks {
			for _, bookmark := range originalCollection.Bookmarks {
				newBookmark := database.Bookmark{
					UserID:      userID,
					URL:         bookmark.URL,
					Title:       bookmark.Title,
					Description: bookmark.Description,
					Tags:        bookmark.Tags,
				}

				if err := tx.Create(&newBookmark).Error; err != nil {
					return fmt.Errorf("failed to create forked bookmark: %w", err)
				}

				// Associate bookmark with forked collection
				if err := tx.Exec("INSERT INTO collection_bookmarks (collection_id, bookmark_id) VALUES (?, ?)",
					forkedCollection.ID, newBookmark.ID).Error; err != nil {
					return fmt.Errorf("failed to associate bookmark with forked collection: %w", err)
				}
			}
		}

		// Create fork record
		fork = &CollectionFork{
			OriginalID:        originalCollectionID,
			ForkedID:          forkedCollection.ID,
			UserID:            userID,
			ForkReason:        request.ForkReason,
			PreserveBookmarks: request.PreserveBookmarks,
			PreserveStructure: request.PreserveStructure,
		}

		if err := tx.Create(fork).Error; err != nil {
			return fmt.Errorf("failed to create fork record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return forkedCollection, nil
}

// AddCollaborator adds a collaborator to a collection
func (s *Service) AddCollaborator(ctx context.Context, userID uint, collectionID uint, request *CollaboratorRequest) (*CollectionCollaborator, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Check if collection exists and user has access
	var collection database.Collection
	if err := s.db.First(&collection, collectionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCollectionNotFound
		}
		return nil, fmt.Errorf("failed to find collection: %w", err)
	}

	if collection.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Find user by email
	var collaboratorUser database.User
	if err := s.db.First(&collaboratorUser, "email = ?", request.Email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", request.Email)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if collaborator already exists
	var existingCollaborator CollectionCollaborator
	if err := s.db.First(&existingCollaborator, "collection_id = ? AND user_id = ?",
		collectionID, collaboratorUser.ID).Error; err == nil {
		return nil, ErrCollaboratorExists
	}

	// Create collaborator
	collaborator := &CollectionCollaborator{
		CollectionID: collectionID,
		UserID:       collaboratorUser.ID,
		InviterID:    userID,
		Permission:   request.Permission,
		Status:       "pending",
		InvitedAt:    time.Now(),
	}

	if err := s.db.Create(collaborator).Error; err != nil {
		return nil, fmt.Errorf("failed to create collaborator: %w", err)
	}

	// TODO: Send invitation email

	return collaborator, nil
}

// AcceptCollaboration accepts a collaboration invitation
func (s *Service) AcceptCollaboration(ctx context.Context, userID uint, collaboratorID uint) error {
	var collaborator CollectionCollaborator
	if err := s.db.First(&collaborator, collaboratorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("collaboration not found")
		}
		return fmt.Errorf("failed to find collaboration: %w", err)
	}

	if collaborator.UserID != userID {
		return ErrUnauthorized
	}

	if collaborator.Status != "pending" {
		return fmt.Errorf("collaboration already %s", collaborator.Status)
	}

	now := time.Now()
	collaborator.Status = "accepted"
	collaborator.AcceptedAt = &now

	if err := s.db.Save(&collaborator).Error; err != nil {
		return fmt.Errorf("failed to accept collaboration: %w", err)
	}

	return nil
}

// RecordActivity records activity on a shared collection
func (s *Service) RecordActivity(ctx context.Context, shareID uint, userID *uint, activityType, ipAddress, userAgent string, metadata map[string]interface{}) error {
	activity := &ShareActivity{
		ShareID:      shareID,
		UserID:       userID,
		ActivityType: activityType,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	// TODO: Serialize metadata to JSON

	if err := s.db.Create(activity).Error; err != nil {
		return fmt.Errorf("failed to record activity: %w", err)
	}

	// Update view count if it's a view activity
	if activityType == "view" {
		if err := s.db.Model(&CollectionShare{}).Where("id = ?", shareID).
			UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			// Log error but don't fail the request
			fmt.Printf("failed to update view count: %v\n", err)
		}
	}

	return nil
}

// GetShareActivity retrieves activity for a share
func (s *Service) GetShareActivity(ctx context.Context, userID uint, shareID uint) ([]ShareActivity, error) {
	// Check if user owns the share
	var share CollectionShare
	if err := s.db.First(&share, shareID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrShareNotFound
		}
		return nil, fmt.Errorf("failed to find share: %w", err)
	}

	if share.UserID != userID {
		return nil, ErrUnauthorized
	}

	var activities []ShareActivity
	if err := s.db.Find(&activities, "share_id = ?", shareID).Error; err != nil {
		return nil, fmt.Errorf("failed to get share activity: %w", err)
	}

	return activities, nil
}

// generateShareToken generates a unique share token
func (s *Service) generateShareToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
