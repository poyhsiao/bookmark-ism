package bookmark

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
)

// Service handles bookmark business logic
type Service struct {
	db *gorm.DB
}

// NewService creates a new bookmark service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateBookmarkRequest represents the request to create a bookmark
type CreateBookmarkRequest struct {
	UserID      uint     `json:"user_id"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Favicon     string   `json:"favicon"`
	Screenshot  string   `json:"screenshot"`
}

// UpdateBookmarkRequest represents the request to update a bookmark
type UpdateBookmarkRequest struct {
	ID          uint     `json:"id"`
	UserID      uint     `json:"user_id"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Favicon     string   `json:"favicon"`
	Screenshot  string   `json:"screenshot"`
}

// ListBookmarksRequest represents the request to list bookmarks
type ListBookmarksRequest struct {
	UserID       uint   `json:"user_id"`
	Search       string `json:"search"`
	Tags         string `json:"tags"`
	CollectionID uint   `json:"collection_id"`
	Status       string `json:"status"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
	SortBy       string `json:"sort_by"`    // created_at, updated_at, title, url
	SortOrder    string `json:"sort_order"` // asc, desc
}

// Create creates a new bookmark
func (s *Service) Create(req CreateBookmarkRequest) (*database.Bookmark, error) {
	// Validate required fields
	if req.URL == "" || req.Title == "" {
		return nil, errors.New("URL and title are required")
	}

	// Validate URL format
	if !isValidURL(req.URL) {
		return nil, errors.New("invalid URL format")
	}

	// Check if user exists
	var user database.User
	if err := s.db.First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to check user: %w", err)
	}

	// Convert tags to JSON
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		tagsJSON = string(tagsBytes)
	}

	// Create bookmark
	bookmark := &database.Bookmark{
		UserID:      req.UserID,
		URL:         req.URL,
		Title:       req.Title,
		Description: req.Description,
		Favicon:     req.Favicon,
		Screenshot:  req.Screenshot,
		Tags:        tagsJSON,
		Status:      "active",
	}

	if err := s.db.Create(bookmark).Error; err != nil {
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}

	return bookmark, nil
}

// GetByID retrieves a bookmark by ID for a specific user
func (s *Service) GetByID(bookmarkID, userID uint) (*database.Bookmark, error) {
	var bookmark database.Bookmark

	err := s.db.Where("id = ? AND user_id = ?", bookmarkID, userID).
		Preload("User").
		Preload("Collections").
		First(&bookmark).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bookmark not found")
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	return &bookmark, nil
}

// Update updates an existing bookmark
func (s *Service) Update(req UpdateBookmarkRequest) (*database.Bookmark, error) {
	// Get existing bookmark
	bookmark, err := s.GetByID(req.ID, req.UserID)
	if err != nil {
		return nil, err
	}

	// Validate URL if provided
	if req.URL != "" && !isValidURL(req.URL) {
		return nil, errors.New("invalid URL format")
	}

	// Update fields if provided
	updates := make(map[string]interface{})

	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Favicon != "" {
		updates["favicon"] = req.Favicon
	}
	if req.Screenshot != "" {
		updates["screenshot"] = req.Screenshot
	}

	// Handle tags
	if req.Tags != nil {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		updates["tags"] = string(tagsBytes)
	}

	// Perform update
	if err := s.db.Model(bookmark).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update bookmark: %w", err)
	}

	// Return updated bookmark
	return s.GetByID(req.ID, req.UserID)
}

// Delete soft deletes a bookmark
func (s *Service) Delete(bookmarkID, userID uint) error {
	// Check if bookmark exists and belongs to user
	_, err := s.GetByID(bookmarkID, userID)
	if err != nil {
		return err
	}

	// Soft delete the bookmark
	if err := s.db.Delete(&database.Bookmark{}, bookmarkID).Error; err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	return nil
}

// List retrieves bookmarks for a user with filtering and pagination
func (s *Service) List(req ListBookmarksRequest) ([]*database.Bookmark, int64, error) {
	query := s.db.Model(&database.Bookmark{}).Where("user_id = ?", req.UserID)

	// Apply filters
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(url) LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Tags != "" {
		// Simple tag search - can be enhanced with JSON queries
		tagTerm := "%" + req.Tags + "%"
		query = query.Where("tags LIKE ?", tagTerm)
	}

	if req.CollectionID > 0 {
		// Join with bookmark_collections table
		query = query.Joins("JOIN bookmark_collections ON bookmarks.id = bookmark_collections.bookmark_id").
			Where("bookmark_collections.collection_id = ?", req.CollectionID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count bookmarks: %w", err)
	}

	// Apply sorting
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Execute query
	var bookmarksData []database.Bookmark
	if err := query.Preload("User").Preload("Collections").Find(&bookmarksData).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list bookmarks: %w", err)
	}

	// Convert to pointer slice
	bookmarks := make([]*database.Bookmark, len(bookmarksData))
	for i := range bookmarksData {
		bookmarks[i] = &bookmarksData[i]
	}

	return bookmarks, total, nil
}

// isValidURL validates if a string is a valid URL
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
