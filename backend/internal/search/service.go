package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/search"

	"github.com/typesense/typesense-go/typesense/api"
)

// Service provides search functionality using Typesense
type Service struct {
	client *search.Client
}

// SearchParams represents advanced search parameters
type SearchParams struct {
	Query       string     `json:"query"`
	UserID      string     `json:"user_id"`
	Tags        []string   `json:"tags,omitempty"`
	Collections []string   `json:"collections,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	SortBy      string     `json:"sort_by,omitempty"`
	SortDesc    bool       `json:"sort_desc,omitempty"`
	Page        int        `json:"page"`
	Limit       int        `json:"limit"`
}

// SearchResult represents search results
type SearchResult struct {
	Bookmarks []BookmarkSearchResult `json:"bookmarks"`
	Total     int                    `json:"total"`
	Page      int                    `json:"page"`
	Limit     int                    `json:"limit"`
	Query     string                 `json:"query"`
}

// BookmarkSearchResult represents a bookmark in search results
type BookmarkSearchResult struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	URL         string              `json:"url"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Tags        []string            `json:"tags"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Highlights  map[string][]string `json:"highlights,omitempty"`
	Score       float64             `json:"score,omitempty"`
}

// CollectionSearchResult represents a collection in search results
type CollectionSearchResult struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	BookmarkCount int                 `json:"bookmark_count"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Highlights    map[string][]string `json:"highlights,omitempty"`
	Score         float64             `json:"score,omitempty"`
}

// SuggestionResult represents search suggestions
type SuggestionResult struct {
	Suggestions []string `json:"suggestions"`
	Query       string   `json:"query"`
}

// NewService creates a new search service
func NewService(cfg config.SearchConfig) (*Service, error) {
	client, err := search.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create search client: %w", err)
	}

	return &Service{
		client: client,
	}, nil
}

// InitializeCollections creates the necessary search collections
func (s *Service) InitializeCollections(ctx context.Context) error {
	// Create bookmarks collection
	if err := s.client.CreateBookmarkCollection(ctx); err != nil {
		// Ignore error if collection already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create bookmarks collection: %w", err)
		}
	}

	// Create collections collection
	if err := s.client.CreateCollectionCollection(ctx); err != nil {
		// Ignore error if collection already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create collections collection: %w", err)
		}
	}

	return nil
}

// IndexBookmark indexes a bookmark in the search engine
func (s *Service) IndexBookmark(ctx context.Context, bookmark *database.Bookmark) error {
	// Parse tags from JSON string
	var tags []string
	if bookmark.Tags != "" {
		// For now, we'll handle tags as a simple array
		// In a real implementation, you'd parse the JSON
		tags = []string{} // TODO: Parse JSON tags
	}

	doc := map[string]interface{}{
		"id":          fmt.Sprintf("%d", bookmark.ID),
		"user_id":     fmt.Sprintf("%d", bookmark.UserID),
		"url":         bookmark.URL,
		"title":       bookmark.Title,
		"description": bookmark.Description,
		"tags":        tags,
		"created_at":  bookmark.CreatedAt.Unix(),
		"updated_at":  bookmark.UpdatedAt.Unix(),
		"save_count":  bookmark.SaveCount,
	}

	return s.client.IndexBookmark(ctx, doc)
}

// UpdateBookmark updates a bookmark in the search engine
func (s *Service) UpdateBookmark(ctx context.Context, bookmark *database.Bookmark) error {
	// Parse tags from JSON string
	var tags []string
	if bookmark.Tags != "" {
		// For now, we'll handle tags as a simple array
		// In a real implementation, you'd parse the JSON
		tags = []string{} // TODO: Parse JSON tags
	}

	doc := map[string]interface{}{
		"id":          fmt.Sprintf("%d", bookmark.ID),
		"user_id":     fmt.Sprintf("%d", bookmark.UserID),
		"url":         bookmark.URL,
		"title":       bookmark.Title,
		"description": bookmark.Description,
		"tags":        tags,
		"created_at":  bookmark.CreatedAt.Unix(),
		"updated_at":  bookmark.UpdatedAt.Unix(),
		"save_count":  bookmark.SaveCount,
	}

	return s.client.UpdateDocument(ctx, "bookmarks", fmt.Sprintf("%d", bookmark.ID), doc)
}

// DeleteBookmark removes a bookmark from the search engine
func (s *Service) DeleteBookmark(ctx context.Context, bookmarkID string) error {
	return s.client.DeleteDocument(ctx, "bookmarks", bookmarkID)
}

// IndexCollection indexes a collection in the search engine
func (s *Service) IndexCollection(ctx context.Context, collection *database.Collection) error {
	doc := map[string]interface{}{
		"id":             fmt.Sprintf("%d", collection.ID),
		"user_id":        fmt.Sprintf("%d", collection.UserID),
		"name":           collection.Name,
		"description":    collection.Description,
		"visibility":     collection.Visibility,
		"created_at":     collection.CreatedAt.Unix(),
		"updated_at":     collection.UpdatedAt.Unix(),
		"bookmark_count": len(collection.Bookmarks),
	}

	return s.client.IndexCollection(ctx, doc)
}

// UpdateCollection updates a collection in the search engine
func (s *Service) UpdateCollection(ctx context.Context, collection *database.Collection) error {
	doc := map[string]interface{}{
		"id":             fmt.Sprintf("%d", collection.ID),
		"user_id":        fmt.Sprintf("%d", collection.UserID),
		"name":           collection.Name,
		"description":    collection.Description,
		"visibility":     collection.Visibility,
		"created_at":     collection.CreatedAt.Unix(),
		"updated_at":     collection.UpdatedAt.Unix(),
		"bookmark_count": len(collection.Bookmarks),
	}

	return s.client.UpdateDocument(ctx, "collections", fmt.Sprintf("%d", collection.ID), doc)
}

// DeleteCollection removes a collection from the search engine
func (s *Service) DeleteCollection(ctx context.Context, collectionID string) error {
	return s.client.DeleteDocument(ctx, "collections", collectionID)
}

// SearchBookmarksBasic performs a basic bookmark search
func (s *Service) SearchBookmarksBasic(ctx context.Context, query, userID string, page, limit int) (*SearchResult, error) {
	params := SearchParams{
		Query:  query,
		UserID: userID,
		Page:   page,
		Limit:  limit,
	}

	return s.SearchBookmarksAdvanced(ctx, params)
}

// SearchBookmarksAdvanced performs an advanced bookmark search
func (s *Service) SearchBookmarksAdvanced(ctx context.Context, params SearchParams) (*SearchResult, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid search parameters: %w", err)
	}

	// Build filter
	filterBy := fmt.Sprintf("user_id:%s", params.UserID)

	// Add tag filters
	if len(params.Tags) > 0 {
		tagFilters := make([]string, len(params.Tags))
		for i, tag := range params.Tags {
			tagFilters[i] = fmt.Sprintf("tags:%s", tag)
		}
		filterBy += " && (" + strings.Join(tagFilters, " || ") + ")"
	}

	// Add date filters
	if params.DateFrom != nil {
		filterBy += fmt.Sprintf(" && created_at:>=%d", params.DateFrom.Unix())
	}
	if params.DateTo != nil {
		filterBy += fmt.Sprintf(" && created_at:<=%d", params.DateTo.Unix())
	}

	// Build sort
	sortBy := "save_count:desc"
	if params.SortBy != "" {
		direction := "asc"
		if params.SortDesc {
			direction = "desc"
		}
		sortBy = fmt.Sprintf("%s:%s", params.SortBy, direction)
	}

	// Prepare search parameters
	queryByWeights := "4,3,2,1"
	highlightFields := "title,description"
	snippetThreshold := 30
	numTypos := "2,1,0"
	minLen1Typo := 4
	minLen2Typo := 7

	searchParams := &api.SearchCollectionParams{
		Q:                params.Query,
		QueryBy:          "title,description,url,tags",
		QueryByWeights:   &queryByWeights,
		FilterBy:         &filterBy,
		SortBy:           &sortBy,
		Page:             &params.Page,
		PerPage:          &params.Limit,
		HighlightFields:  &highlightFields,
		SnippetThreshold: &snippetThreshold,
		NumTypos:         &numTypos,
		MinLen1typo:      &minLen1Typo,
		MinLen2typo:      &minLen2Typo,
	}

	// Perform search
	result, err := s.client.Search(ctx, "bookmarks", searchParams)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	var bookmarks []BookmarkSearchResult
	if result.Hits != nil {
		bookmarks = make([]BookmarkSearchResult, 0, len(*result.Hits))
		for _, hit := range *result.Hits {
			bookmark, err := s.convertToBookmarkResult(hit)
			if err != nil {
				continue // Skip invalid results
			}
			bookmarks = append(bookmarks, bookmark)
		}
	}

	total := 0
	if result.Found != nil {
		total = *result.Found
	}

	return &SearchResult{
		Bookmarks: bookmarks,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		Query:     params.Query,
	}, nil
}

// SearchCollections searches for collections
func (s *Service) SearchCollections(ctx context.Context, query, userID string, page, limit int) (*api.SearchResult, error) {
	filterBy := fmt.Sprintf("user_id:%s", userID)
	sortBy := "bookmark_count:desc"

	searchParams := &api.SearchCollectionParams{
		Q:        query,
		QueryBy:  "name,description",
		FilterBy: &filterBy,
		Page:     &page,
		PerPage:  &limit,
		SortBy:   &sortBy,
	}

	return s.client.Search(ctx, "collections", searchParams)
}

// GetSuggestions returns search suggestions based on partial query
func (s *Service) GetSuggestions(ctx context.Context, query, userID string, limit int) (*SuggestionResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	filterBy := fmt.Sprintf("user_id:%s", userID)
	page := 1
	sortBy := "save_count:desc"

	searchParams := &api.SearchCollectionParams{
		Q:        query,
		QueryBy:  "title,tags",
		FilterBy: &filterBy,
		Page:     &page,
		PerPage:  &limit,
		SortBy:   &sortBy,
	}

	result, err := s.client.Search(ctx, "bookmarks", searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	// Extract unique suggestions from results
	suggestionSet := make(map[string]bool)
	suggestions := make([]string, 0, limit)

	if result.Hits != nil {
		for _, hit := range *result.Hits {
			if hit.Document != nil {
				doc := *hit.Document
				// Add title suggestions
				if title, ok := doc["title"].(string); ok && title != "" {
					if !suggestionSet[title] && len(suggestions) < limit {
						suggestionSet[title] = true
						suggestions = append(suggestions, title)
					}
				}

				// Add tag suggestions
				if tags, ok := doc["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok && tagStr != "" {
							if !suggestionSet[tagStr] && len(suggestions) < limit {
								suggestionSet[tagStr] = true
								suggestions = append(suggestions, tagStr)
							}
						}
					}
				}
			}
		}
	}

	return &SuggestionResult{
		Suggestions: suggestions,
		Query:       query,
	}, nil
}

// HealthCheck checks if the search service is healthy
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.client.HealthCheck(ctx)
}

// convertToBookmarkResult converts a Typesense hit to BookmarkSearchResult
func (s *Service) convertToBookmarkResult(hit api.SearchResultHit) (BookmarkSearchResult, error) {
	if hit.Document == nil {
		return BookmarkSearchResult{}, fmt.Errorf("document is nil")
	}

	doc := *hit.Document

	result := BookmarkSearchResult{}

	// Extract basic fields
	if id, ok := doc["id"].(string); ok {
		result.ID = id
	}
	if userID, ok := doc["user_id"].(string); ok {
		result.UserID = userID
	}
	if url, ok := doc["url"].(string); ok {
		result.URL = url
	}
	if title, ok := doc["title"].(string); ok {
		result.Title = title
	}
	if description, ok := doc["description"].(string); ok {
		result.Description = description
	}

	// Extract tags
	if tags, ok := doc["tags"].([]interface{}); ok {
		result.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				result.Tags = append(result.Tags, tagStr)
			}
		}
	}

	// Extract timestamps
	if createdAt, ok := doc["created_at"].(float64); ok {
		result.CreatedAt = time.Unix(int64(createdAt), 0)
	}
	if updatedAt, ok := doc["updated_at"].(float64); ok {
		result.UpdatedAt = time.Unix(int64(updatedAt), 0)
	}

	// Extract highlights
	if hit.Highlights != nil {
		result.Highlights = make(map[string][]string)
		for _, highlight := range *hit.Highlights {
			if highlight.Field != nil && highlight.Snippets != nil && len(*highlight.Snippets) > 0 {
				result.Highlights[*highlight.Field] = *highlight.Snippets
			}
		}
	}

	// Extract score
	if hit.TextMatch != nil {
		result.Score = float64(*hit.TextMatch)
	}

	return result, nil
}

// Validate validates search parameters
func (p *SearchParams) Validate() error {
	if p.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if p.Page <= 0 {
		return fmt.Errorf("page must be greater than 0")
	}

	if p.Limit <= 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	if p.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	// Validate sort field
	if p.SortBy != "" {
		validSortFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
			"save_count": true,
		}
		if !validSortFields[p.SortBy] {
			return fmt.Errorf("invalid sort field: %s", p.SortBy)
		}
	}

	return nil
}
