package search

import (
	"net/http"
	"strconv"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handlers provides HTTP handlers for search functionality
type Handlers struct {
	service *Service
}

// NewHandlers creates new search handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRoutes registers search routes
func (h *Handlers) RegisterRoutes(router *gin.RouterGroup) {
	search := router.Group("/search")
	{
		// Search endpoints
		search.GET("/bookmarks", h.SearchBookmarksBasic)
		search.POST("/bookmarks/advanced", h.SearchBookmarksAdvanced)
		search.GET("/collections", h.SearchCollections)
		search.GET("/suggestions", h.GetSuggestions)

		// Index management endpoints
		search.POST("/index/bookmark", h.IndexBookmark)
		search.PUT("/index/bookmark/:id", h.UpdateBookmark)
		search.DELETE("/index/bookmark/:id", h.DeleteBookmark)
		search.POST("/index/collection", h.IndexCollection)
		search.PUT("/index/collection/:id", h.UpdateCollection)
		search.DELETE("/index/collection/:id", h.DeleteCollection)

		// System endpoints
		search.GET("/health", h.HealthCheck)
		search.POST("/initialize", h.InitializeCollections)
	}
}

// SearchBookmarksBasic handles basic bookmark search
func (h *Handlers) SearchBookmarksBasic(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	query := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid page parameter", nil)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid limit parameter (must be 1-100)", nil)
		return
	}

	result, err := h.service.SearchBookmarksBasic(c.Request.Context(), query, userID.(string), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_FAILED", "Search failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Search completed successfully")
}

// SearchBookmarksAdvanced handles advanced bookmark search
func (h *Handlers) SearchBookmarksAdvanced(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var params SearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set user ID from context
	params.UserID = userID.(string)

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	result, err := h.service.SearchBookmarksAdvanced(c.Request.Context(), params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_FAILED", "Advanced search failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Advanced search completed successfully")
}

// SearchCollections handles collection search
func (h *Handlers) SearchCollections(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	query := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid page parameter", nil)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid limit parameter (must be 1-100)", nil)
		return
	}

	result, err := h.service.SearchCollections(c.Request.Context(), query, userID.(string), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_FAILED", "Collection search failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Collection search completed successfully")
}

// GetSuggestions handles search suggestions
func (h *Handlers) GetSuggestions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Query parameter 'q' is required", nil)
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 20 {
		limit = 5
	}

	result, err := h.service.GetSuggestions(c.Request.Context(), query, userID.(string), limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SUGGESTIONS_FAILED", "Failed to get suggestions", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, result, "Suggestions retrieved successfully")
}

// IndexBookmark handles bookmark indexing
func (h *Handlers) IndexBookmark(c *gin.Context) {
	var bookmark database.Bookmark
	if err := c.ShouldBindJSON(&bookmark); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid bookmark data", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.service.IndexBookmark(c.Request.Context(), &bookmark); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INDEX_FAILED", "Failed to index bookmark", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Bookmark indexed successfully"}, "Bookmark indexed successfully")
}

// UpdateBookmark handles bookmark update in search index
func (h *Handlers) UpdateBookmark(c *gin.Context) {
	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Bookmark ID is required", nil)
		return
	}

	var bookmark database.Bookmark
	if err := c.ShouldBindJSON(&bookmark); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid bookmark data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Parse bookmark ID from URL parameter
	bookmarkIDUint, err := strconv.ParseUint(bookmarkID, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid bookmark ID format", nil)
		return
	}
	bookmark.ID = uint(bookmarkIDUint)

	if err := h.service.UpdateBookmark(c.Request.Context(), &bookmark); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update bookmark in search index", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Bookmark updated in search index successfully"}, "Bookmark updated in search index successfully")
}

// DeleteBookmark handles bookmark deletion from search index
func (h *Handlers) DeleteBookmark(c *gin.Context) {
	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Bookmark ID is required", nil)
		return
	}

	if err := h.service.DeleteBookmark(c.Request.Context(), bookmarkID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete bookmark from search index", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Bookmark deleted from search index successfully"}, "Bookmark deleted from search index successfully")
}

// IndexCollection handles collection indexing
func (h *Handlers) IndexCollection(c *gin.Context) {
	var collection database.Collection
	if err := c.ShouldBindJSON(&collection); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collection data", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.service.IndexCollection(c.Request.Context(), &collection); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INDEX_FAILED", "Failed to index collection", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Collection indexed successfully"}, "Collection indexed successfully")
}

// UpdateCollection handles collection update in search index
func (h *Handlers) UpdateCollection(c *gin.Context) {
	collectionID := c.Param("id")
	if collectionID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Collection ID is required", nil)
		return
	}

	var collection database.Collection
	if err := c.ShouldBindJSON(&collection); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collection data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Parse collection ID from URL parameter
	collectionIDUint, err := strconv.ParseUint(collectionID, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid collection ID format", nil)
		return
	}
	collection.ID = uint(collectionIDUint)

	if err := h.service.UpdateCollection(c.Request.Context(), &collection); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update collection in search index", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Collection updated in search index successfully"}, "Collection updated in search index successfully")
}

// DeleteCollection handles collection deletion from search index
func (h *Handlers) DeleteCollection(c *gin.Context) {
	collectionID := c.Param("id")
	if collectionID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Collection ID is required", nil)
		return
	}

	if err := h.service.DeleteCollection(c.Request.Context(), collectionID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete collection from search index", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Collection deleted from search index successfully"}, "Collection deleted from search index successfully")
}

// HealthCheck handles search service health check
func (h *Handlers) HealthCheck(c *gin.Context) {
	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Search service is not healthy", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{
		"status":    "healthy",
		"service":   "search",
		"timestamp": time.Now().UTC(),
	}, "Search service is healthy")
}

// InitializeCollections handles search collections initialization
func (h *Handlers) InitializeCollections(c *gin.Context) {
	if err := h.service.InitializeCollections(c.Request.Context()); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INITIALIZATION_FAILED", "Failed to initialize search collections", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Search collections initialized successfully"}, "Search collections initialized successfully")
}
