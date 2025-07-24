package bookmark

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bookmark-sync-service/backend/pkg/utils"
)

// Handlers handles HTTP requests for bookmark operations
type Handlers struct {
	service *Service
}

// NewHandlers creates a new bookmark handlers instance
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRoutes registers bookmark routes
func (h *Handlers) RegisterRoutes(router *gin.RouterGroup) {
	bookmarks := router.Group("/bookmarks")
	{
		bookmarks.POST("", h.CreateBookmark)
		bookmarks.GET("", h.ListBookmarksHandler)
		bookmarks.GET("/:id", h.GetBookmark)
		bookmarks.PUT("/:id", h.UpdateBookmark)
		bookmarks.DELETE("/:id", h.DeleteBookmark)
	}
}

// CreateBookmark creates a new bookmark
func (h *Handlers) CreateBookmark(c *gin.Context) {
	var req CreateBookmarkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format", nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	req.UserID = userID.(uint)

	bookmark, err := h.service.Create(req)
	if err != nil {
		if err.Error() == "URL and title are required" || err.Error() == "invalid URL format" {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if err.Error() == "user not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create bookmark", nil)
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Bookmark created successfully",
		Data:    bookmark,
	})
}

// GetBookmark retrieves a bookmark by ID
func (h *Handlers) GetBookmark(c *gin.Context) {
	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid bookmark ID", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	bookmark, err := h.service.GetByID(uint(bookmarkID), userID.(uint))
	if err != nil {
		if err.Error() == "bookmark not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get bookmark", nil)
		return
	}

	utils.SuccessResponse(c, bookmark, "Bookmark retrieved successfully")
}

// UpdateBookmark updates an existing bookmark
func (h *Handlers) UpdateBookmark(c *gin.Context) {
	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid bookmark ID", nil)
		return
	}

	var req UpdateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	req.ID = uint(bookmarkID)
	req.UserID = userID.(uint)

	bookmark, err := h.service.Update(req)
	if err != nil {
		if err.Error() == "bookmark not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		if err.Error() == "invalid URL format" {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update bookmark", nil)
		return
	}

	utils.SuccessResponse(c, bookmark, "Bookmark updated successfully")
}

// DeleteBookmark deletes a bookmark
func (h *Handlers) DeleteBookmark(c *gin.Context) {
	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid bookmark ID", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	err = h.service.Delete(uint(bookmarkID), userID.(uint))
	if err != nil {
		if err.Error() == "bookmark not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete bookmark", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Bookmark deleted successfully")
}

// ListBookmarksHandler lists bookmarks with filtering and pagination
func (h *Handlers) ListBookmarksHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Parse query parameters
	req := ListBookmarksRequest{
		UserID: userID.(uint),
		Search: c.Query("search"),
		Tags:   c.Query("tags"),
		Status: c.Query("status"),
	}

	// Parse collection ID
	if collectionIDStr := c.Query("collection_id"); collectionIDStr != "" {
		collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
			return
		}
		req.CollectionID = uint(collectionID)
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit parameter", nil)
			return
		}
		req.Limit = limit
	} else {
		req.Limit = 20 // Default limit
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid offset parameter", nil)
			return
		}
		req.Offset = offset
	}

	// Parse sort parameters
	req.SortBy = c.Query("sort_by")
	req.SortOrder = c.Query("sort_order")

	bookmarks, total, err := h.service.List(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list bookmarks", nil)
		return
	}

	response := map[string]interface{}{
		"bookmarks": bookmarks,
		"total":     total,
		"limit":     req.Limit,
		"offset":    req.Offset,
	}

	utils.SuccessResponse(c, response, "Bookmarks retrieved successfully")
}
