package search

import (
	"fmt"
	"time"
)

// FacetedSearchParams represents parameters for faceted search
type FacetedSearchParams struct {
	Query     string            `json:"query"`
	UserID    string            `json:"user_id"`
	FacetBy   []string          `json:"facet_by,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
	MaxFacets int               `json:"max_facets,omitempty"`
	Page      int               `json:"page"`
	Limit     int               `json:"limit"`
}

// FacetedSearchResult represents faceted search results
type FacetedSearchResult struct {
	Bookmarks []BookmarkSearchResult  `json:"bookmarks"`
	Facets    map[string][]FacetValue `json:"facets"`
	Total     int                     `json:"total"`
	Page      int                     `json:"page"`
	Limit     int                     `json:"limit"`
	Query     string                  `json:"query"`
}

// FacetValue represents a facet value with count
type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// SemanticSearchParams represents parameters for semantic search
type SemanticSearchParams struct {
	Query     string   `json:"query"`
	UserID    string   `json:"user_id"`
	Intent    string   `json:"intent,omitempty"`
	Context   []string `json:"context,omitempty"`
	Threshold float64  `json:"threshold,omitempty"`
	Page      int      `json:"page"`
	Limit     int      `json:"limit"`
}

// AutoCompleteResult represents auto-complete suggestions
type AutoCompleteResult struct {
	Suggestions []AutoCompleteSuggestion `json:"suggestions"`
	Query       string                   `json:"query"`
}

// AutoCompleteSuggestion represents a single auto-complete suggestion
type AutoCompleteSuggestion struct {
	Text  string `json:"text"`
	Type  string `json:"type"` // "title", "tag", "url", "description"
	Count int    `json:"count"`
}

// ClusteredSearchResult represents search results organized in clusters
type ClusteredSearchResult struct {
	Clusters []SearchCluster `json:"clusters"`
	Total    int             `json:"total"`
}

// SearchCluster represents a cluster of related search results
type SearchCluster struct {
	Name      string                 `json:"name"`
	Bookmarks []BookmarkSearchResult `json:"bookmarks"`
	Score     float64                `json:"score"`
	Tags      []string               `json:"tags,omitempty"`
}

// SavedSearch represents a saved search query
type SavedSearch struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Name       string                 `json:"name"`
	Query      string                 `json:"query"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	LastUsedAt *time.Time             `json:"last_used_at,omitempty"`
	UseCount   int                    `json:"use_count"`
}

// SearchHistoryEntry represents a search history entry
type SearchHistoryEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Query     string    `json:"query"`
	Results   int       `json:"results"`
	CreatedAt time.Time `json:"created_at"`
}

// SearchHistoryResult represents search history results
type SearchHistoryResult struct {
	Entries []SearchHistoryEntry `json:"entries"`
	Total   int                  `json:"total"`
}

// Validate validates faceted search parameters
func (p *FacetedSearchParams) Validate() error {
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

	if p.MaxFacets <= 0 {
		p.MaxFacets = 10
	}

	if p.MaxFacets > 50 {
		return fmt.Errorf("max_facets cannot exceed 50")
	}

	// Validate facet fields
	validFacetFields := map[string]bool{
		"tags":       true,
		"created_at": true,
		"updated_at": true,
		"domain":     true,
	}

	for _, field := range p.FacetBy {
		if !validFacetFields[field] {
			return fmt.Errorf("invalid facet field: %s", field)
		}
	}

	return nil
}

// Validate validates semantic search parameters
func (p *SemanticSearchParams) Validate() error {
	if p.Query == "" {
		return fmt.Errorf("query is required")
	}

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

	if p.Threshold < 0 || p.Threshold > 1 {
		p.Threshold = 0.5 // Default threshold
	}

	return nil
}

// Validate validates saved search
func (s *SavedSearch) Validate() error {
	if s.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if s.Name == "" {
		return fmt.Errorf("name is required")
	}

	if s.Query == "" {
		return fmt.Errorf("query is required")
	}

	if len(s.Name) > 100 {
		return fmt.Errorf("name cannot exceed 100 characters")
	}

	if len(s.Query) > 500 {
		return fmt.Errorf("query cannot exceed 500 characters")
	}

	return nil
}
