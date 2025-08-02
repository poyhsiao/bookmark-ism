package search

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/redis"

	"github.com/typesense/typesense-go/typesense/api"
	"gorm.io/gorm"
)

// AdvancedSearchClient interface for advanced search operations
type AdvancedSearchClient interface {
	// Basic search operations
	CreateBookmarkCollection(ctx context.Context) error
	CreateCollectionCollection(ctx context.Context) error
	IndexBookmark(ctx context.Context, doc map[string]interface{}) error
	IndexCollection(ctx context.Context, doc map[string]interface{}) error
	UpdateDocument(ctx context.Context, collection, id string, doc map[string]interface{}) error
	DeleteDocument(ctx context.Context, collection, id string) error
	Search(ctx context.Context, collection string, params *api.SearchCollectionParams) (*api.SearchResult, error)
	HealthCheck(ctx context.Context) error

	// Advanced search operations
	FacetedSearch(ctx context.Context, collection string, params *FacetedSearchParams) (*FacetedSearchResult, error)
	SemanticSearch(ctx context.Context, collection string, params *SemanticSearchParams) (*SearchResult, error)
	GetAutoComplete(ctx context.Context, collection, query string, limit int) (*AutoCompleteResult, error)
	ClusterResults(ctx context.Context, results []BookmarkSearchResult) (*ClusteredSearchResult, error)
}

// AdvancedService provides advanced search functionality
type AdvancedService struct {
	*Service
	db          *gorm.DB
	redisClient *redis.Client
}

// NewAdvancedService creates a new advanced search service
func NewAdvancedService(service *Service, db *gorm.DB, redisClient *redis.Client) *AdvancedService {
	return &AdvancedService{
		Service:     service,
		db:          db,
		redisClient: redisClient,
	}
}

// FacetedSearch performs faceted search with aggregated facets
func (s *AdvancedService) FacetedSearch(ctx context.Context, params FacetedSearchParams) (*FacetedSearchResult, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Build filter
	filterBy := fmt.Sprintf("user_id:%s", params.UserID)

	// Add custom filters
	for field, value := range params.Filters {
		filterBy += fmt.Sprintf(" && %s:%s", field, value)
	}

	// Build facet by string
	facetBy := strings.Join(params.FacetBy, ",")

	// Prepare search parameters with faceting
	queryByWeights := "4,3,2,1"
	highlightFields := "title,description"
	snippetThreshold := 30
	numTypos := "2,1,0"
	minLen1Typo := 4
	minLen2Typo := 7
	maxFacetValues := params.MaxFacets

	searchParams := &api.SearchCollectionParams{
		Q:                params.Query,
		QueryBy:          "title,description,url,tags",
		QueryByWeights:   &queryByWeights,
		FilterBy:         &filterBy,
		FacetBy:          &facetBy,
		MaxFacetValues:   &maxFacetValues,
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
		return nil, fmt.Errorf("faceted search failed: %w", err)
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

	// Convert facets
	facets := make(map[string][]FacetValue)
	if result.FacetCounts != nil {
		for _, facetCount := range *result.FacetCounts {
			if facetCount.FieldName != nil && facetCount.Counts != nil {
				fieldName := *facetCount.FieldName
				facetValues := make([]FacetValue, 0, len(*facetCount.Counts))

				for _, count := range *facetCount.Counts {
					if count.Value != nil && count.Count != nil {
						facetValues = append(facetValues, FacetValue{
							Value: *count.Value,
							Count: *count.Count,
						})
					}
				}

				// Sort facet values by count (descending)
				sort.Slice(facetValues, func(i, j int) bool {
					return facetValues[i].Count > facetValues[j].Count
				})

				facets[fieldName] = facetValues
			}
		}
	}

	total := 0
	if result.Found != nil {
		total = *result.Found
	}

	return &FacetedSearchResult{
		Bookmarks: bookmarks,
		Facets:    facets,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		Query:     params.Query,
	}, nil
}

// SemanticSearch performs semantic search with natural language processing
func (s *AdvancedService) SemanticSearch(ctx context.Context, params SemanticSearchParams) (*SearchResult, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Enhance query with semantic understanding
	enhancedQuery := s.enhanceQuerySemantics(params.Query, params.Intent, params.Context)

	// Build advanced search parameters
	filterBy := fmt.Sprintf("user_id:%s", params.UserID)

	// Use semantic ranking
	sortBy := "_text_match:desc,save_count:desc"
	queryByWeights := "5,4,3,2,1" // Higher weight for semantic matching
	highlightFields := "title,description"
	snippetThreshold := 30
	numTypos := "1,0" // Stricter typo tolerance for semantic search
	minLen1Typo := 6
	minLen2Typo := 8

	searchParams := &api.SearchCollectionParams{
		Q:                enhancedQuery,
		QueryBy:          "title,description,url,tags,content",
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
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Convert results with semantic scoring
	var bookmarks []BookmarkSearchResult
	if result.Hits != nil {
		bookmarks = make([]BookmarkSearchResult, 0, len(*result.Hits))
		for _, hit := range *result.Hits {
			bookmark, err := s.convertToBookmarkResult(hit)
			if err != nil {
				continue // Skip invalid results
			}

			// Apply semantic scoring adjustments
			bookmark.Score = s.calculateSemanticScore(bookmark, params)
			bookmarks = append(bookmarks, bookmark)
		}

		// Re-sort by semantic score
		sort.Slice(bookmarks, func(i, j int) bool {
			return bookmarks[i].Score > bookmarks[j].Score
		})
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

// GetAutoComplete provides intelligent auto-complete suggestions
func (s *AdvancedService) GetAutoComplete(ctx context.Context, query, userID string, limit int) (*AutoCompleteResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if limit <= 0 || limit > 20 {
		limit = 5
	}

	// Get suggestions from multiple sources
	suggestions := make([]AutoCompleteSuggestion, 0, limit)

	// 1. Get title suggestions
	titleSuggestions, err := s.getTitleSuggestions(ctx, query, userID, limit/2)
	if err == nil {
		suggestions = append(suggestions, titleSuggestions...)
	}

	// 2. Get tag suggestions
	tagSuggestions, err := s.getTagSuggestions(ctx, query, userID, limit/2)
	if err == nil {
		suggestions = append(suggestions, tagSuggestions...)
	}

	// 3. Get URL domain suggestions
	domainSuggestions, err := s.getDomainSuggestions(ctx, query, userID, limit/4)
	if err == nil {
		suggestions = append(suggestions, domainSuggestions...)
	}

	// Sort by relevance and count
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Count != suggestions[j].Count {
			return suggestions[i].Count > suggestions[j].Count
		}
		return len(suggestions[i].Text) < len(suggestions[j].Text)
	})

	// Limit results
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return &AutoCompleteResult{
		Suggestions: suggestions,
		Query:       query,
	}, nil
}

// ClusterResults organizes search results into semantic clusters
func (s *AdvancedService) ClusterResults(ctx context.Context, results []BookmarkSearchResult) (*ClusteredSearchResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to cluster")
	}

	// Simple clustering based on tags and domains
	clusters := make(map[string]*SearchCluster)

	for _, result := range results {
		clusterKey := s.determineClusterKey(result)

		if cluster, exists := clusters[clusterKey]; exists {
			cluster.Bookmarks = append(cluster.Bookmarks, result)
			cluster.Score += result.Score * 0.1 // Accumulate cluster score
		} else {
			clusters[clusterKey] = &SearchCluster{
				Name:      s.generateClusterName(clusterKey, result),
				Bookmarks: []BookmarkSearchResult{result},
				Score:     result.Score,
				Tags:      s.extractClusterTags(result),
			}
		}
	}

	// Convert to slice and sort by score
	clusterSlice := make([]SearchCluster, 0, len(clusters))
	for _, cluster := range clusters {
		clusterSlice = append(clusterSlice, *cluster)
	}

	sort.Slice(clusterSlice, func(i, j int) bool {
		return clusterSlice[i].Score > clusterSlice[j].Score
	})

	return &ClusteredSearchResult{
		Clusters: clusterSlice,
		Total:    len(results),
	}, nil
}

// SaveSearch saves a search query for later use
func (s *AdvancedService) SaveSearch(ctx context.Context, savedSearch *SavedSearch) error {
	if err := savedSearch.Validate(); err != nil {
		return fmt.Errorf("invalid saved search: %w", err)
	}

	// Set timestamps
	now := time.Now()
	savedSearch.CreatedAt = now
	savedSearch.UpdatedAt = now

	// Save to database
	if err := s.db.Create(savedSearch).Error; err != nil {
		return fmt.Errorf("failed to save search: %w", err)
	}

	return nil
}

// GetSavedSearches retrieves saved searches for a user
func (s *AdvancedService) GetSavedSearches(ctx context.Context, userID string) ([]SavedSearch, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	var savedSearches []SavedSearch
	if err := s.db.Where("user_id = ?", userID).
		Order("last_used_at DESC, created_at DESC").
		Find(&savedSearches).Error; err != nil {
		return nil, fmt.Errorf("failed to get saved searches: %w", err)
	}

	return savedSearches, nil
}

// DeleteSavedSearch deletes a saved search
func (s *AdvancedService) DeleteSavedSearch(ctx context.Context, searchID, userID string) error {
	if searchID == "" {
		return fmt.Errorf("search_id is required")
	}

	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	result := s.db.Where("id = ? AND user_id = ?", searchID, userID).Delete(&SavedSearch{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete saved search: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("saved search not found")
	}

	return nil
}

// RecordSearchHistory records a search in the user's history
func (s *AdvancedService) RecordSearchHistory(ctx context.Context, userID, query string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	if query == "" {
		return fmt.Errorf("query is required")
	}

	// Use Redis for search history (temporary storage)
	key := fmt.Sprintf("search_history:%s", userID)

	historyEntry := SearchHistoryEntry{
		UserID:    userID,
		Query:     query,
		CreatedAt: time.Now(),
	}

	entryJSON, err := json.Marshal(historyEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal history entry: %w", err)
	}

	// Add to Redis list (keep last 100 searches)
	if err := s.redisClient.Client.LPush(ctx, key, string(entryJSON)).Err(); err != nil {
		return fmt.Errorf("failed to record search history: %w", err)
	}

	// Trim to keep only last 100 entries
	if err := s.redisClient.Client.LTrim(ctx, key, 0, 99).Err(); err != nil {
		return fmt.Errorf("failed to trim search history: %w", err)
	}

	// Set expiration (30 days)
	if err := s.redisClient.Client.Expire(ctx, key, 30*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set search history expiration: %w", err)
	}

	return nil
}

// GetSearchHistory retrieves search history for a user
func (s *AdvancedService) GetSearchHistory(ctx context.Context, userID string, limit int) (*SearchHistoryResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	key := fmt.Sprintf("search_history:%s", userID)

	// Get from Redis
	entries, err := s.redisClient.Client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get search history: %w", err)
	}

	historyEntries := make([]SearchHistoryEntry, 0, len(entries))
	for _, entryStr := range entries {
		var entry SearchHistoryEntry
		if err := json.Unmarshal([]byte(entryStr), &entry); err != nil {
			continue // Skip invalid entries
		}
		historyEntries = append(historyEntries, entry)
	}

	return &SearchHistoryResult{
		Entries: historyEntries,
		Total:   len(historyEntries),
	}, nil
}

// ClearSearchHistory clears search history for a user
func (s *AdvancedService) ClearSearchHistory(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	key := fmt.Sprintf("search_history:%s", userID)

	if err := s.redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to clear search history: %w", err)
	}

	return nil
}

// Helper methods

func (s *AdvancedService) enhanceQuerySemantics(query, intent string, context []string) string {
	// Simple semantic enhancement
	enhanced := query

	// Add intent-based keywords
	if intent != "" {
		switch intent {
		case "learning":
			enhanced += " tutorial guide learn"
		case "reference":
			enhanced += " documentation reference api"
		case "news":
			enhanced += " news article blog"
		}
	}

	// Add context keywords
	if len(context) > 0 {
		enhanced += " " + strings.Join(context, " ")
	}

	return enhanced
}

func (s *AdvancedService) calculateSemanticScore(bookmark BookmarkSearchResult, params SemanticSearchParams) float64 {
	score := bookmark.Score

	// Boost score based on semantic relevance
	if params.Intent != "" {
		if s.matchesIntent(bookmark, params.Intent) {
			score *= 1.2
		}
	}

	// Boost score based on context
	if len(params.Context) > 0 {
		contextMatches := s.countContextMatches(bookmark, params.Context)
		score *= (1.0 + float64(contextMatches)*0.1)
	}

	return score
}

func (s *AdvancedService) matchesIntent(bookmark BookmarkSearchResult, intent string) bool {
	text := strings.ToLower(bookmark.Title + " " + bookmark.Description)

	switch intent {
	case "learning":
		return strings.Contains(text, "tutorial") || strings.Contains(text, "guide") || strings.Contains(text, "learn")
	case "reference":
		return strings.Contains(text, "documentation") || strings.Contains(text, "reference") || strings.Contains(text, "api")
	case "news":
		return strings.Contains(text, "news") || strings.Contains(text, "article") || strings.Contains(text, "blog")
	}

	return false
}

func (s *AdvancedService) countContextMatches(bookmark BookmarkSearchResult, context []string) int {
	text := strings.ToLower(bookmark.Title + " " + bookmark.Description + " " + strings.Join(bookmark.Tags, " "))
	matches := 0

	for _, ctx := range context {
		if strings.Contains(text, strings.ToLower(ctx)) {
			matches++
		}
	}

	return matches
}

func (s *AdvancedService) getTitleSuggestions(ctx context.Context, query, userID string, limit int) ([]AutoCompleteSuggestion, error) {
	// Simple implementation - in production, this would use more sophisticated matching
	var bookmarks []database.Bookmark
	if err := s.db.Where("user_id = ? AND title ILIKE ?", userID, "%"+query+"%").
		Limit(limit).Find(&bookmarks).Error; err != nil {
		return nil, err
	}

	suggestions := make([]AutoCompleteSuggestion, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		suggestions = append(suggestions, AutoCompleteSuggestion{
			Text:  bookmark.Title,
			Type:  "title",
			Count: 1, // In production, this would be actual usage count
		})
	}

	return suggestions, nil
}

func (s *AdvancedService) getTagSuggestions(ctx context.Context, query, userID string, limit int) ([]AutoCompleteSuggestion, error) {
	// This would need a proper tag index in production
	var bookmarks []database.Bookmark
	if err := s.db.Where("user_id = ? AND tags ILIKE ?", userID, "%"+query+"%").
		Limit(limit * 2).Find(&bookmarks).Error; err != nil {
		return nil, err
	}

	tagCounts := make(map[string]int)
	for _, bookmark := range bookmarks {
		// Simple tag extraction - in production, parse JSON properly
		if strings.Contains(bookmark.Tags, query) {
			tagCounts[query]++
		}
	}

	suggestions := make([]AutoCompleteSuggestion, 0, len(tagCounts))
	for tag, count := range tagCounts {
		suggestions = append(suggestions, AutoCompleteSuggestion{
			Text:  tag,
			Type:  "tag",
			Count: count,
		})
	}

	return suggestions, nil
}

func (s *AdvancedService) getDomainSuggestions(ctx context.Context, query, userID string, limit int) ([]AutoCompleteSuggestion, error) {
	// Extract domain suggestions from URLs
	var bookmarks []database.Bookmark
	if err := s.db.Where("user_id = ? AND url ILIKE ?", userID, "%"+query+"%").
		Limit(limit * 2).Find(&bookmarks).Error; err != nil {
		return nil, err
	}

	domainCounts := make(map[string]int)
	for _, bookmark := range bookmarks {
		if domain := s.extractDomain(bookmark.URL); domain != "" && strings.Contains(domain, query) {
			domainCounts[domain]++
		}
	}

	suggestions := make([]AutoCompleteSuggestion, 0, len(domainCounts))
	for domain, count := range domainCounts {
		suggestions = append(suggestions, AutoCompleteSuggestion{
			Text:  domain,
			Type:  "domain",
			Count: count,
		})
	}

	return suggestions, nil
}

func (s *AdvancedService) extractDomain(url string) string {
	// Simple domain extraction
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}

func (s *AdvancedService) determineClusterKey(result BookmarkSearchResult) string {
	// Simple clustering by domain or primary tag
	domain := s.extractDomain(result.URL)
	if domain != "" {
		return "domain:" + domain
	}

	if len(result.Tags) > 0 {
		return "tag:" + result.Tags[0]
	}

	return "misc"
}

func (s *AdvancedService) generateClusterName(key string, result BookmarkSearchResult) string {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) == 2 {
		switch parts[0] {
		case "domain":
			return fmt.Sprintf("From %s", parts[1])
		case "tag":
			return fmt.Sprintf("Tagged: %s", parts[1])
		}
	}
	return "Other"
}

func (s *AdvancedService) extractClusterTags(result BookmarkSearchResult) []string {
	if len(result.Tags) > 3 {
		return result.Tags[:3]
	}
	return result.Tags
}
