package content

import (
	"net/http"
	"strconv"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/middleware"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for content analysis
type Handler struct {
	service *Service
	cfg     *config.Config
}

// NewHandler creates a new content analysis handler
func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

// RegisterRoutes registers the content analysis routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	content := router.Group("/content")
	content.Use(middleware.AuthMiddleware(&h.cfg.JWT))
	{
		content.POST("/analyze", h.AnalyzeURL)
		content.POST("/suggest-tags", h.SuggestTags)
		content.POST("/detect-duplicates", h.DetectDuplicates)
		content.POST("/categorize", h.CategorizeContent)
		content.POST("/bookmarks/:id/analyze", h.AnalyzeBookmarkContent)
	}
}

// AnalyzeURL analyzes a URL and returns comprehensive analysis
// @Summary Analyze URL content
// @Description Performs comprehensive content analysis including tag suggestions, categorization, and duplicate detection
// @Tags content
// @Accept json
// @Produce json
// @Param request body AnalysisRequest true "Analysis request"
// @Success 200 {object} AnalysisResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/content/analyze [post]
func (h *Handler) AnalyzeURL(c *gin.Context) {
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", nil)
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Convert string to uint
	userID := uint(1) // Default fallback
	if userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userID = uint(id)
		}
	}

	// Override user ID from request if provided (for admin users)
	if req.UserID != 0 {
		userID = req.UserID
	}

	// Perform analysis
	result, err := h.service.AnalyzeURL(c.Request.Context(), req.URL, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "ANALYSIS_FAILED", "Failed to analyze URL", nil)
		return
	}

	utils.SuccessResponse(c, result, "URL analyzed successfully")
}

// SuggestTags suggests tags for a bookmark
// @Summary Suggest tags for bookmark
// @Description Suggests relevant tags based on bookmark content analysis
// @Tags content
// @Accept json
// @Produce json
// @Param request body TagSuggestionRequest true "Tag suggestion request"
// @Success 200 {object} map[string][]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/content/suggest-tags [post]
func (h *Handler) SuggestTags(c *gin.Context) {
	var req TagSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", nil)
		return
	}

	// Validate user access to bookmark
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get tag suggestions
	tags, err := h.service.SuggestTagsForBookmark(c.Request.Context(), req.BookmarkID, req.URL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "TAG_SUGGESTION_FAILED", "Failed to suggest tags", nil)
		return
	}

	result := gin.H{
		"bookmark_id":    req.BookmarkID,
		"suggested_tags": tags,
	}

	utils.SuccessResponse(c, result, "Tags suggested successfully")
}

// DetectDuplicates detects potential duplicate bookmarks
// @Summary Detect duplicate bookmarks
// @Description Finds potential duplicate bookmarks for a user based on content similarity
// @Tags content
// @Accept json
// @Produce json
// @Param request body DuplicateDetectionRequest true "Duplicate detection request"
// @Success 200 {object} map[string][]*DuplicateMatch
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/content/detect-duplicates [post]
func (h *Handler) DetectDuplicates(c *gin.Context) {
	var req DuplicateDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", nil)
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Convert string to uint
	userID := uint(1) // Default fallback
	if userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userID = uint(id)
		}
	}

	// Override user ID from request if provided
	if req.UserID != 0 {
		userID = req.UserID
	}

	// Detect duplicates
	duplicates, err := h.service.DetectDuplicateBookmarks(c.Request.Context(), req.URL, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DUPLICATE_DETECTION_FAILED", "Failed to detect duplicates", nil)
		return
	}

	result := gin.H{
		"url":        req.URL,
		"user_id":    userID,
		"duplicates": duplicates,
	}

	utils.SuccessResponse(c, result, "Duplicates detected successfully")
}

// CategorizeContent categorizes content based on analysis
// @Summary Categorize content
// @Description Categorizes content into predefined categories based on content analysis
// @Tags content
// @Accept json
// @Produce json
// @Param request body CategoryRequest true "Category request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/content/categorize [post]
func (h *Handler) CategorizeContent(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", nil)
		return
	}

	// Categorize content
	category, err := h.service.CategorizeBookmark(c.Request.Context(), req.URL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CATEGORIZATION_FAILED", "Failed to categorize content", nil)
		return
	}

	result := gin.H{
		"url":      req.URL,
		"category": category,
	}

	utils.SuccessResponse(c, result, "Content categorized successfully")
}

// AnalyzeBookmarkContent analyzes content for an existing bookmark
// @Summary Analyze existing bookmark content
// @Description Analyzes content for an existing bookmark by ID
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Bookmark ID"
// @Success 200 {object} AnalysisResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/content/bookmarks/{id}/analyze [post]
func (h *Handler) AnalyzeBookmarkContent(c *gin.Context) {
	// Get bookmark ID from path
	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BOOKMARK_ID", "Invalid bookmark ID", nil)
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Convert string to uint
	userID := uint(1) // Default fallback
	if userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userID = uint(id)
		}
	}

	// TODO: Fetch bookmark URL from database
	// For now, we'll expect the URL to be provided in the request body
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_URL", "URL is required", nil)
		return
	}

	// Perform analysis
	result, err := h.service.AnalyzeURL(c.Request.Context(), req.URL, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "BOOKMARK_ANALYSIS_FAILED", "Failed to analyze bookmark content", nil)
		return
	}

	// Add bookmark ID to result
	response := gin.H{
		"bookmark_id": uint(bookmarkID),
		"analysis":    result,
	}

	utils.SuccessResponse(c, response, "Bookmark content analyzed successfully")
}
