package collection

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
)

// Service handles collection business logic
type Service struct {
	db *gorm.DB
}

// NewService creates a new collection service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateCollectionRequest represents a request to create a collection
type CreateCollectionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Visibility  string `json:"visibility" binding:"required,oneof=private public shared"`
}

// UpdateCollectionRequest represents a request to update a collection
type UpdateCollectionRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ParentID    *uint   `json:"parent_id,omitempty"`
	Visibility  *string `json:"visibility,omitempty" binding:"omitempty,oneof=private public shared"`
}

// ListCollectionsParams represents parameters for listing collections
type ListCollectionsParams struct {
	Page       int    `form:"page,default=1" binding:"min=1"`
	Limit      int    `form:"limit,default=20" binding:"min=1,max=100"`
	Search     string `form:"search"`
	Visibility string `form:"visibility" binding:"omitempty,oneof=private public shared"`
	ParentID   *uint  `form:"parent_id"`
	SortBy     string `form:"sort_by,default=created_at" binding:"oneof=created_at updated_at name"`
	SortOrder  string `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

// ListCollectionsResult represents the result of listing collections
type ListCollectionsResult struct {
	Collections []database.Collection `json:"collections"`
	Total       int64                 `json:"total"`
	Page        int                   `json:"page"`
	Limit       int                   `json:"limit"`
	TotalPages  int                   `json:"total_pages"`
}

// GetCollectionBookmarksParams represents parameters for getting collection bookmarks
type GetCollectionBookmarksParams struct {
	Page      int    `form:"page,default=1" binding:"min=1"`
	Limit     int    `form:"limit,default=20" binding:"min=1,max=100"`
	Search    string `form:"search"`
	SortBy    string `form:"sort_by,default=created_at" binding:"oneof=created_at updated_at title url"`
	SortOrder string `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

// GetCollectionBookmarksResult represents the result of getting collection bookmarks
type GetCollectionBookmarksResult struct {
	Bookmarks  []database.Bookmark `json:"bookmarks"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

// Create creates a new collection
func (s *Service) Create(userID uint, req CreateCollectionRequest) (*database.Collection, error) {
	// Validate request
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}

	if req.Visibility != "private" && req.Visibility != "public" && req.Visibility != "shared" {
		return nil, errors.New("invalid visibility")
	}

	// Validate parent collection if specified
	if req.ParentID != nil {
		var parent database.Collection
		if err := s.db.Where("id = ? AND user_id = ?", *req.ParentID, userID).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent collection not found")
			}
			return nil, fmt.Errorf("failed to validate parent collection: %w", err)
		}
	}

	// Generate share link
	shareLink, err := s.generateShareLink()
	if err != nil {
		return nil, fmt.Errorf("failed to generate share link: %w", err)
	}

	// Create collection
	collection := &database.Collection{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		ParentID:    req.ParentID,
		Visibility:  req.Visibility,
		ShareLink:   shareLink,
	}

	if err := s.db.Create(collection).Error; err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return collection, nil
}

// GetByID retrieves a collection by ID
func (s *Service) GetByID(userID, id uint) (*database.Collection, error) {
	var collection database.Collection

	query := s.db.Where("id = ?", id)

	// For private collections, ensure user ownership
	// For public collections, allow access by anyone
	// For shared collections, allow access by anyone with the link (handled in handlers)
	query = query.Where("user_id = ? OR visibility = ?", userID, "public")

	if err := query.Preload("User").Preload("Parent").First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	return &collection, nil
}

// List retrieves collections with filtering and pagination
func (s *Service) List(userID uint, params ListCollectionsParams) (*ListCollectionsResult, error) {
	var collections []database.Collection
	var total int64

	// Set default values if not provided
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Build base query
	query := s.db.Model(&database.Collection{}).Where("user_id = ?", userID)

	// Apply filters
	if params.Search != "" {
		searchTerm := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if params.Visibility != "" {
		query = query.Where("visibility = ?", params.Visibility)
	}

	if params.ParentID != nil {
		query = query.Where("parent_id = ?", *params.ParentID)
	} else {
		// By default, only show root collections (no parent)
		query = query.Where("parent_id IS NULL")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count collections: %w", err)
	}

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", params.SortBy, strings.ToUpper(params.SortOrder))
	query = query.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Execute query with preloading
	if err := query.Preload("User").Preload("Parent").Find(&collections).Error; err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	// Calculate total pages
	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))

	return &ListCollectionsResult{
		Collections: collections,
		Total:       total,
		Page:        params.Page,
		Limit:       params.Limit,
		TotalPages:  totalPages,
	}, nil
}

// Update updates a collection
func (s *Service) Update(userID, id uint, req UpdateCollectionRequest) (*database.Collection, error) {
	// Get existing collection
	var collection database.Collection
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Validate updates
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, errors.New("name cannot be empty")
		}
		collection.Name = strings.TrimSpace(*req.Name)
	}

	if req.Description != nil {
		collection.Description = *req.Description
	}

	if req.Color != nil {
		collection.Color = *req.Color
	}

	if req.Icon != nil {
		collection.Icon = *req.Icon
	}

	if req.Visibility != nil {
		if *req.Visibility != "private" && *req.Visibility != "public" && *req.Visibility != "shared" {
			return nil, errors.New("invalid visibility")
		}
		collection.Visibility = *req.Visibility
	}

	if req.ParentID != nil {
		// Validate parent collection
		var parent database.Collection
		if err := s.db.Where("id = ? AND user_id = ?", *req.ParentID, userID).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent collection not found")
			}
			return nil, fmt.Errorf("failed to validate parent collection: %w", err)
		}
		collection.ParentID = req.ParentID
	}

	// Save updates
	if err := s.db.Save(&collection).Error; err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	return &collection, nil
}

// Delete soft deletes a collection
func (s *Service) Delete(userID, id uint) error {
	// Check if collection exists and belongs to user
	var collection database.Collection
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("collection not found")
		}
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Soft delete the collection
	if err := s.db.Delete(&collection).Error; err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return nil
}

// AddBookmark adds a bookmark to a collection
func (s *Service) AddBookmark(userID, collectionID, bookmarkID uint) error {
	// Verify collection exists and belongs to user
	var collection database.Collection
	if err := s.db.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("collection not found")
		}
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Verify bookmark exists and belongs to user
	var bookmark database.Bookmark
	if err := s.db.Where("id = ? AND user_id = ?", bookmarkID, userID).First(&bookmark).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("bookmark not found")
		}
		return fmt.Errorf("failed to get bookmark: %w", err)
	}

	// Add bookmark to collection (GORM handles duplicates automatically)
	if err := s.db.Model(&collection).Association("Bookmarks").Append(&bookmark); err != nil {
		return fmt.Errorf("failed to add bookmark to collection: %w", err)
	}

	return nil
}

// RemoveBookmark removes a bookmark from a collection
func (s *Service) RemoveBookmark(userID, collectionID, bookmarkID uint) error {
	// Verify collection exists and belongs to user
	var collection database.Collection
	if err := s.db.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("collection not found")
		}
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Remove bookmark from collection
	var bookmark database.Bookmark
	bookmark.ID = bookmarkID
	if err := s.db.Model(&collection).Association("Bookmarks").Delete(&bookmark); err != nil {
		return fmt.Errorf("failed to remove bookmark from collection: %w", err)
	}

	return nil
}

// GetBookmarks retrieves bookmarks in a collection with filtering and pagination
func (s *Service) GetBookmarks(userID, collectionID uint, params GetCollectionBookmarksParams) (*GetCollectionBookmarksResult, error) {
	// Set default values if not provided
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Verify collection exists and user has access
	var collection database.Collection
	query := s.db.Where("id = ?", collectionID)
	query = query.Where("user_id = ? OR visibility = ?", userID, "public")

	if err := query.First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var bookmarks []database.Bookmark
	var total int64

	// Build query for bookmarks in this collection
	bookmarkQuery := s.db.Model(&database.Bookmark{}).
		Joins("JOIN bookmark_collections ON bookmarks.id = bookmark_collections.bookmark_id").
		Where("bookmark_collections.collection_id = ?", collectionID)

	// Apply search filter
	if params.Search != "" {
		searchTerm := "%" + strings.ToLower(params.Search) + "%"
		bookmarkQuery = bookmarkQuery.Where(
			"LOWER(bookmarks.title) LIKE ? OR LOWER(bookmarks.description) LIKE ? OR LOWER(bookmarks.url) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Count total records
	if err := bookmarkQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count bookmarks: %w", err)
	}

	// Apply sorting
	orderClause := fmt.Sprintf("bookmarks.%s %s", params.SortBy, strings.ToUpper(params.SortOrder))
	bookmarkQuery = bookmarkQuery.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	bookmarkQuery = bookmarkQuery.Offset(offset).Limit(params.Limit)

	// Execute query
	if err := bookmarkQuery.Preload("User").Find(&bookmarks).Error; err != nil {
		return nil, fmt.Errorf("failed to get collection bookmarks: %w", err)
	}

	// Calculate total pages
	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))

	return &GetCollectionBookmarksResult{
		Bookmarks:  bookmarks,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// generateShareLink generates a unique share link for a collection
func (s *Service) generateShareLink() (string, error) {
	const maxAttempts = 5

	for i := 0; i < maxAttempts; i++ {
		// Generate random bytes
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Convert to hex string
		shareLink := hex.EncodeToString(bytes)

		// Check if link already exists
		var count int64
		if err := s.db.Model(&database.Collection{}).Where("share_link = ?", shareLink).Count(&count).Error; err != nil {
			return "", fmt.Errorf("failed to check share link uniqueness: %w", err)
		}

		if count == 0 {
			return shareLink, nil
		}
	}

	return "", errors.New("failed to generate unique share link after multiple attempts")
}
