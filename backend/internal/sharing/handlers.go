package sharing

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bookmark-sync-service/backend/pkg/middleware"
	"bookmark-sync-service/backend/pkg/utils"
)

// Handler represents the sharing HTTP handler
type Handler struct {
	service *Service
}

// NewHandler creates a new sharing handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateShare creates a new collection share
// @Summary Create a new collection share
// @Description Create a new share for a collection with specified permissions
// @Tags sharing
// @Accept json
// @Produce json
// @Param request body CreateShareRequest true "Share creation request"
// @Success 201 {object} ShareResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/shares [post]
func (h *Handler) CreateShare(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	var request CreateShareRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	share, err := h.service.CreateShare(c.Request.Context(), uint(userID), &request)
	if err != nil {
		switch err {
		case ErrCollectionNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "collection_not_found", "collection not found", nil)
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		case ErrInvalidCollectionID, ErrInvalidShareType, ErrInvalidPermission:
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request parameters", map[string]interface{}{"error": err.Error()})
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to create share", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "share created successfully",
		Data:    share,
	})
}

// GetShare retrieves a share by token
// @Summary Get share by token
// @Description Retrieve a shared collection by its share token
// @Tags sharing
// @Produce json
// @Param token path string true "Share token"
// @Param password query string false "Password for protected shares"
// @Success 200 {object} ShareResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 410 {object} utils.ErrorResponse "Share expired"
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/shares/{token} [get]
func (h *Handler) GetShare(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "missing_token", "share token is required", nil)
		return
	}

	share, err := h.service.GetShareByToken(c.Request.Context(), token)
	if err != nil {
		switch err {
		case ErrShareNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "share_not_found", "share not found", nil)
		case ErrShareExpired:
			utils.ErrorResponse(c, http.StatusGone, "share_expired", "share has expired", nil)
		case ErrShareInactive:
			utils.ErrorResponse(c, http.StatusGone, "share_inactive", "share is inactive", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get share", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	// Check password if required
	if share.Password != "" {
		password := c.Query("password")
		if password != share.Password { // TODO: Use proper password hashing
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid_password", "invalid password", nil)
			return
		}
	}

	// Record view activity
	userIDStr := middleware.GetUserID(c)
	var userIDPtr *uint
	if userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(userID)
			userIDPtr = &uid
		}
	}

	if err := h.service.RecordActivity(c.Request.Context(), share.ID, userIDPtr, "view",
		c.ClientIP(), c.GetHeader("User-Agent"), nil); err != nil {
		// Log error but don't fail the request
		// TODO: Use proper logging
	}

	response := share.ToResponse("http://localhost:3000") // TODO: Get base URL from config
	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "share retrieved successfully",
		Data:    response,
	})
}

// UpdateShare updates an existing share
// @Summary Update share
// @Description Update an existing collection share
// @Tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Share ID"
// @Param request body UpdateShareRequest true "Share update request"
// @Success 200 {object} ShareResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/shares/{id} [put]
func (h *Handler) UpdateShare(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseUint(shareIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_share_id", "invalid share ID", nil)
		return
	}

	var request UpdateShareRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	share, err := h.service.UpdateShare(c.Request.Context(), uint(userID), uint(shareID), &request)
	if err != nil {
		switch err {
		case ErrShareNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "share_not_found", "share not found", nil)
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		case ErrInvalidShareType, ErrInvalidPermission:
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request parameters", map[string]interface{}{"error": err.Error()})
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to update share", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "share updated successfully",
		Data:    share,
	})
}

// DeleteShare deletes a share
// @Summary Delete share
// @Description Delete a collection share
// @Tags sharing
// @Param id path int true "Share ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/shares/{id} [delete]
func (h *Handler) DeleteShare(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseUint(shareIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_share_id", "invalid share ID", nil)
		return
	}

	if err := h.service.DeleteShare(c.Request.Context(), uint(userID), uint(shareID)); err != nil {
		switch err {
		case ErrShareNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "share_not_found", "share not found", nil)
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to delete share", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "share deleted successfully",
	})
}

// GetUserShares retrieves all shares for the authenticated user
// @Summary Get user shares
// @Description Retrieve all shares created by the authenticated user
// @Tags sharing
// @Produce json
// @Success 200 {array} ShareResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/shares [get]
func (h *Handler) GetUserShares(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	shares, err := h.service.GetUserShares(c.Request.Context(), uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get user shares", map[string]interface{}{"error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]ShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = *share.ToResponse("http://localhost:3000") // TODO: Get base URL from config
	}

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "user shares retrieved successfully",
		Data:    responses,
	})
}

// GetCollectionShares retrieves all shares for a specific collection
// @Summary Get collection shares
// @Description Retrieve all shares for a specific collection
// @Tags sharing
// @Produce json
// @Param id path int true "Collection ID"
// @Success 200 {array} ShareResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id}/shares [get]
func (h *Handler) GetCollectionShares(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_collection_id", "invalid collection ID", nil)
		return
	}

	shares, err := h.service.GetCollectionShares(c.Request.Context(), uint(userID), uint(collectionID))
	if err != nil {
		switch err {
		case ErrCollectionNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "collection_not_found", "collection not found", nil)
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get collection shares", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	// Convert to response format
	responses := make([]ShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = *share.ToResponse("http://localhost:3000") // TODO: Get base URL from config
	}

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "collection shares retrieved successfully",
		Data:    responses,
	})
}

// ForkCollection creates a fork of a shared collection
// @Summary Fork collection
// @Description Create a fork of a shared collection
// @Tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Original Collection ID"
// @Param request body ForkRequest true "Fork request"
// @Success 201 {object} database.Collection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id}/fork [post]
func (h *Handler) ForkCollection(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_collection_id", "invalid collection ID", nil)
		return
	}

	var request ForkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	forkedCollection, err := h.service.ForkCollection(c.Request.Context(), uint(userID), uint(collectionID), &request)
	if err != nil {
		switch err {
		case ErrCollectionNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "collection_not_found", "collection not found", nil)
		case ErrCannotForkOwnCollection:
			utils.ErrorResponse(c, http.StatusBadRequest, "cannot_fork_own", "cannot fork own collection", nil)
		case ErrForkNotAllowed:
			utils.ErrorResponse(c, http.StatusForbidden, "fork_not_allowed", "fork not allowed", nil)
		case ErrInvalidName:
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request parameters", map[string]interface{}{"error": err.Error()})
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to fork collection", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "collection forked successfully",
		Data:    forkedCollection,
	})
}

// AddCollaborator adds a collaborator to a collection
// @Summary Add collaborator
// @Description Add a collaborator to a collection
// @Tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param request body CollaboratorRequest true "Collaborator request"
// @Success 201 {object} CollectionCollaborator
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse "Collaborator already exists"
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collections/{id}/collaborators [post]
func (h *Handler) AddCollaborator(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	collectionIDStr := c.Param("id")
	collectionID, err := strconv.ParseUint(collectionIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_collection_id", "invalid collection ID", nil)
		return
	}

	var request CollaboratorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	collaborator, err := h.service.AddCollaborator(c.Request.Context(), uint(userID), uint(collectionID), &request)
	if err != nil {
		switch err {
		case ErrCollectionNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "collection_not_found", "collection not found", nil)
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		case ErrCollaboratorExists:
			utils.ErrorResponse(c, http.StatusConflict, "collaborator_exists", "collaborator already exists", nil)
		case ErrInvalidEmail, ErrInvalidPermission:
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "invalid request parameters", map[string]interface{}{"error": err.Error()})
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to add collaborator", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "collaborator added successfully",
		Data:    collaborator,
	})
}

// AcceptCollaboration accepts a collaboration invitation
// @Summary Accept collaboration
// @Description Accept a collaboration invitation
// @Tags sharing
// @Param id path int true "Collaborator ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/collaborations/{id}/accept [post]
func (h *Handler) AcceptCollaboration(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	collaboratorIDStr := c.Param("id")
	collaboratorID, err := strconv.ParseUint(collaboratorIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_collaborator_id", "invalid collaborator ID", nil)
		return
	}

	if err := h.service.AcceptCollaboration(c.Request.Context(), uint(userID), uint(collaboratorID)); err != nil {
		switch err {
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to accept collaboration", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "collaboration accepted successfully",
	})
}

// GetShareActivity retrieves activity for a share
// @Summary Get share activity
// @Description Retrieve activity logs for a share
// @Tags sharing
// @Produce json
// @Param id path int true "Share ID"
// @Success 200 {array} ShareActivity
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/shares/{id}/activity [get]
func (h *Handler) GetShareActivity(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "user not authenticated", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "invalid user ID", nil)
		return
	}

	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseUint(shareIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_share_id", "invalid share ID", nil)
		return
	}

	activities, err := h.service.GetShareActivity(c.Request.Context(), uint(userID), uint(shareID))
	if err != nil {
		switch err {
		case ErrShareNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "share_not_found", "share not found", nil)
		case ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, "unauthorized", "unauthorized access", nil)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get share activity", map[string]interface{}{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "share activity retrieved successfully",
		Data:    activities,
	})
}
