package collection

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bookmark-sync-service/backend/pkg/utils"
)

// Handler handles HTTP requests for collections
type Handler struct {
	service *Service
}

// NewHandler creates a new collection handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers collection routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	collections := router.Group("/collections")
	{
		collections.POST("", h.CreateCollection)
		collections.GET("", h.ListCollections)
		collections.GET("/:id", h.GetCollection)
		collections.PUT("/:id", h.UpdateCollection)
		collections.DELETE("/:id", h.DeleteCollection)

		// Bookmark management within collections
		collections.POST("/:id/bookmarks/:bookmark_id", h.AddBookmarkToCollection)
		collections.DELETE("/:id/bookmarks/:bookmark_id", h.RemoveBookmarkFromCollection)
		collections.GET("/:id/bookmarks", h.GetCollectionBookmarks)
	}
}

// CreateCollection creates a new collection
// @Summary Create a new collection
// @Description Create a new bookmark collection
// @Tags collections
// @Accept json
// @Produce json
// @Param collection body CreateCollectionRequest true "Collection data"
// @Success 201 {object} database.Collection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections [post]
func (h *Handler) CreateCollection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", nil)
		return
	}

	collection, err := h.service.Create(userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "CREATE_ERROR", "Failed to create collection", nil)
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Collection created successfully",
		Data:    collection,
	})
}

// ListCollections lists collections with filtering and pagination
// @Summary List collections
// @Description Get a paginated list of collections with optional filtering
// @Tags collections
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search term"
// @Param visibility query string false "Filter by visibility" Enums(private, public, shared)
// @Param parent_id query int false "Filter by parent collection ID"
// @Param sort_by query string false "Sort field" default(created_at) Enums(created_at, updated_at, name)
// @Param sort_order query string false "Sort order" default(desc) Enums(asc, desc)
// @Success 200 {object} ListCollectionsResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections [get]
func (h *Handler) ListCollections(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var params ListCollectionsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", nil)
		return
	}

	result, err := h.service.List(userID.(uint), params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list collections", nil)
		return
	}

	utils.SuccessResponse(c, result, "Collections retrieved successfully")
}

// GetCollection retrieves a collection by ID
// @Summary Get a collection
// @Description Get a collection by its ID
// @Tags collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Success 200 {object} database.Collection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id} [get]
func (h *Handler) GetCollection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
		return
	}

	collection, err := h.service.GetByID(userID.(uint), uint(id))
	if err != nil {
		if err.Error() == "collection not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Collection not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get collection", nil)
		return
	}

	utils.SuccessResponse(c, collection, "Collection retrieved successfully")
}

// UpdateCollection updates a collection
// @Summary Update a collection
// @Description Update an existing collection
// @Tags collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param collection body UpdateCollectionRequest true "Updated collection data"
// @Success 200 {object} database.Collection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id} [put]
func (h *Handler) UpdateCollection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
		return
	}

	var req UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", nil)
		return
	}

	collection, err := h.service.Update(userID.(uint), uint(id), req)
	if err != nil {
		if err.Error() == "collection not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Collection not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "UPDATE_ERROR", "Failed to update collection", nil)
		return
	}

	utils.SuccessResponse(c, collection, "Collection updated successfully")
}

// DeleteCollection deletes a collection
// @Summary Delete a collection
// @Description Soft delete a collection
// @Tags collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id} [delete]
func (h *Handler) DeleteCollection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
		return
	}

	err = h.service.Delete(userID.(uint), uint(id))
	if err != nil {
		if err.Error() == "collection not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Collection not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete collection", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Collection deleted successfully")
}

// AddBookmarkToCollection adds a bookmark to a collection
// @Summary Add bookmark to collection
// @Description Add an existing bookmark to a collection
// @Tags collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param bookmark_id path int true "Bookmark ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id}/bookmarks/{bookmark_id} [post]
func (h *Handler) AddBookmarkToCollection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
		return
	}

	bookmarkIDStr := c.Param("bookmark_id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid bookmark ID", nil)
		return
	}

	err = h.service.AddBookmark(userID.(uint), uint(collectionID), uint(bookmarkID))
	if err != nil {
		if err.Error() == "collection not found" || err.Error() == "bookmark not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "ADD_ERROR", "Failed to add bookmark to collection", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Bookmark added to collection successfully")
}

// RemoveBookmarkFromCollection removes a bookmark from a collection
// @Summary Remove bookmark from collection
// @Description Remove a bookmark from a collection
// @Tags collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param bookmark_id path int true "Bookmark ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id}/bookmarks/{bookmark_id} [delete]
func (h *Handler) RemoveBookmarkFromCollection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
		return
	}

	bookmarkIDStr := c.Param("bookmark_id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid bookmark ID", nil)
		return
	}

	err = h.service.RemoveBookmark(userID.(uint), uint(collectionID), uint(bookmarkID))
	if err != nil {
		if err.Error() == "collection not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Collection not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "REMOVE_ERROR", "Failed to remove bookmark from collection", nil)
		return
	}

	utils.SuccessResponse(c, nil, "Bookmark removed from collection successfully")
}

// GetCollectionBookmarks retrieves bookmarks in a collection
// @Summary Get collection bookmarks
// @Description Get a paginated list of bookmarks in a collection
// @Tags collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field" default(created_at) Enums(created_at, updated_at, title, url)
// @Param sort_order query string false "Sort order" default(desc) Enums(asc, desc)
// @Success 200 {object} GetCollectionBookmarksResult
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id}/bookmarks [get]
func (h *Handler) GetCollectionBookmarks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid collection ID", nil)
		return
	}

	var params GetCollectionBookmarksParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", nil)
		return
	}

	result, err := h.service.GetBookmarks(userID.(uint), uint(collectionID), params)
	if err != nil {
		if err.Error() == "collection not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Collection not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get collection bookmarks", nil)
		return
	}

	utils.SuccessResponse(c, result, "Collection bookmarks retrieved successfully")
}
