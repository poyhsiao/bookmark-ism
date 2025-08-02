package search

import (
	"net/http"
	"strconv"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AdvancedHandlers provides HTTP handlers for advanced search functionality
type AdvancedHandlers struct {
	service *AdvancedService
}

// NewAdvancedHandlers creates new advanced search handlers
func NewAdvancedHandlers(service *AdvancedService) *AdvancedHandlers {
	return &AdvancedHandlers{
		service: service,
	}
}

// FacetedSearch handles faceted search requests
// @Summary Perform faceted search
// @Description Search bookmarks with faceted results and aggregations
// @Tags search
// @Accept json
// @Produce json
// @Param request body FacetedSearchParams true "Faceted search parameters"
// @Success 200 {object} FacetedSearchResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/faceted [post]
func (h *AdvancedHandlers) FacetedSearch(c *gin.Context) {
	var params FacetedSearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request parameters", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}
	params.UserID = userID.(string)

	result, err := h.service.FacetedSearch(c.Request.Context(), params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_ERROR", "Faceted search failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Faceted search completed successfully")
}

// SemanticSearch handles semantic search requests
// @Summary Perform semantic search
// @Description Search bookmarks using natural language processing and semantic understanding
// @Tags search
// @Accept json
// @Produce json
// @Param request body SemanticSearchParams true "Semantic search parameters"
// @Success 200 {object} SearchResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/semantic [post]
func (h *AdvancedHandlers) SemanticSearch(c *gin.Context) {
	var params SemanticSearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request parameters", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}
	params.UserID = userID.(string)

	result, err := h.service.SemanticSearch(c.Request.Context(), params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_ERROR", "Semantic search failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Semantic search completed successfully")
}

// GetAutoComplete handles auto-complete requests
// @Summary Get auto-complete suggestions
// @Description Get intelligent auto-complete suggestions for search queries
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Maximum number of suggestions" default(5)
// @Success 200 {object} AutoCompleteResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/autocomplete [get]
func (h *AdvancedHandlers) GetAutoComplete(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Query parameter 'q' is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Parse limit
	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	result, err := h.service.GetAutoComplete(c.Request.Context(), query, userID.(string), limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_ERROR", "Auto-complete failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Auto-complete suggestions retrieved successfully")
}

// ClusterResults handles search result clustering requests
// @Summary Cluster search results
// @Description Organize search results into semantic clusters
// @Tags search
// @Accept json
// @Produce json
// @Param request body []BookmarkSearchResult true "Search results to cluster"
// @Success 200 {object} ClusteredSearchResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/cluster [post]
func (h *AdvancedHandlers) ClusterResults(c *gin.Context) {
	var results []BookmarkSearchResult
	if err := c.ShouldBindJSON(&results); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request parameters", map[string]interface{}{"error": err.Error()})
		return
	}

	clusteredResult, err := h.service.ClusterResults(c.Request.Context(), results)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_ERROR", "Result clustering failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, clusteredResult, "Search results clustered successfully")
}

// SaveSearch handles saving search queries
// @Summary Save a search query
// @Description Save a search query for later use
// @Tags search
// @Accept json
// @Produce json
// @Param request body SavedSearch true "Search to save"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/saved [post]
func (h *AdvancedHandlers) SaveSearch(c *gin.Context) {
	var savedSearch SavedSearch
	if err := c.ShouldBindJSON(&savedSearch); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request parameters", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}
	savedSearch.UserID = userID.(string)

	if err := h.service.SaveSearch(c.Request.Context(), &savedSearch); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SAVE_ERROR", "Failed to save search", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Search saved successfully", "id": savedSearch.ID}, "Search saved successfully")
}

// GetSavedSearches handles retrieving saved searches
// @Summary Get saved searches
// @Description Retrieve all saved searches for the current user
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {array} SavedSearch
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/saved [get]
func (h *AdvancedHandlers) GetSavedSearches(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	savedSearches, err := h.service.GetSavedSearches(c.Request.Context(), userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to get saved searches", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, savedSearches, "Saved searches retrieved successfully")
}

// DeleteSavedSearch handles deleting saved searches
// @Summary Delete a saved search
// @Description Delete a saved search by ID
// @Tags search
// @Accept json
// @Produce json
// @Param id path string true "Saved search ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/saved/{id} [delete]
func (h *AdvancedHandlers) DeleteSavedSearch(c *gin.Context) {
	searchID := c.Param("id")
	if searchID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Search ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	if err := h.service.DeleteSavedSearch(c.Request.Context(), searchID, userID.(string)); err != nil {
		if err.Error() == "saved search not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Saved search not found", map[string]interface{}{"search_id": searchID})
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete saved search", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Saved search deleted successfully"}, "Saved search deleted successfully")
}

// GetSearchHistory handles retrieving search history
// @Summary Get search history
// @Description Retrieve search history for the current user
// @Tags search
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of entries" default(20)
// @Success 200 {object} SearchHistoryResult
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/history [get]
func (h *AdvancedHandlers) GetSearchHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Parse limit
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	result, err := h.service.GetSearchHistory(c.Request.Context(), userID.(string), limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to get search history", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Search history retrieved successfully")
}

// ClearSearchHistory handles clearing search history
// @Summary Clear search history
// @Description Clear all search history for the current user
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/history [delete]
func (h *AdvancedHandlers) ClearSearchHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	if err := h.service.ClearSearchHistory(c.Request.Context(), userID.(string)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CLEAR_ERROR", "Failed to clear search history", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Search history cleared successfully"}, "Search history cleared successfully")
}

// RecordSearch handles recording search queries in history
// @Summary Record search in history
// @Description Record a search query in the user's search history
// @Tags search
// @Accept json
// @Produce json
// @Param request body gin.H true "Search query" example({"query": "golang tutorial"})
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/search/history [post]
func (h *AdvancedHandlers) RecordSearch(c *gin.Context) {
	var request struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request parameters", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	if err := h.service.RecordSearchHistory(c.Request.Context(), userID.(string), request.Query); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "RECORD_ERROR", "Failed to record search", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Search recorded successfully"}, "Search recorded successfully")
}

// RegisterAdvancedRoutes registers advanced search routes
func RegisterAdvancedRoutes(router *gin.RouterGroup, handlers *AdvancedHandlers) {
	search := router.Group("/search")
	{
		// Advanced search endpoints
		search.POST("/faceted", handlers.FacetedSearch)
		search.POST("/semantic", handlers.SemanticSearch)
		search.GET("/autocomplete", handlers.GetAutoComplete)
		search.POST("/cluster", handlers.ClusterResults)

		// Saved searches
		search.POST("/saved", handlers.SaveSearch)
		search.GET("/saved", handlers.GetSavedSearches)
		search.DELETE("/saved/:id", handlers.DeleteSavedSearch)

		// Search history
		search.GET("/history", handlers.GetSearchHistory)
		search.POST("/history", handlers.RecordSearch)
		search.DELETE("/history", handlers.ClearSearchHistory)
	}
}
