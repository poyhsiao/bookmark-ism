package sync

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles sync-related HTTP requests
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new sync handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetSyncState handles GET /api/v1/sync/state
func (h *Handler) GetSyncState(c *gin.Context) {
	userID := c.GetString("user_id")
	deviceID := c.Query("device_id")

	if deviceID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "device_id is required", nil)
		return
	}

	state, err := h.service.GetSyncState(c.Request.Context(), userID, deviceID)
	if err != nil {
		h.logger.Error("Failed to get sync state", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to get sync state", nil)
		return
	}

	utils.SuccessResponse(c, state, "Sync state retrieved successfully")
}

// GetDeltaSync handles GET /api/v1/sync/delta
func (h *Handler) GetDeltaSync(c *gin.Context) {
	userID := c.GetString("user_id")
	deviceID := c.Query("device_id")
	lastSyncTimeStr := c.Query("last_sync_time")

	if deviceID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "device_id is required", nil)
		return
	}

	var lastSyncTime time.Time
	if lastSyncTimeStr != "" {
		timestamp, err := strconv.ParseInt(lastSyncTimeStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid last_sync_time format", nil)
			return
		}
		lastSyncTime = time.Unix(timestamp, 0)
	} else {
		// Default to 24 hours ago if not provided
		lastSyncTime = time.Now().Add(-24 * time.Hour)
	}

	delta, err := h.service.GetDeltaSync(c.Request.Context(), userID, deviceID, lastSyncTime)
	if err != nil {
		h.logger.Error("Failed to get delta sync", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to get delta sync", nil)
		return
	}

	utils.SuccessResponse(c, delta, "Delta sync retrieved successfully")
}

// CreateSyncEvent handles POST /api/v1/sync/events
func (h *Handler) CreateSyncEvent(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Type       SyncEventType          `json:"type" binding:"required"`
		ResourceID string                 `json:"resource_id" binding:"required"`
		Action     string                 `json:"action" binding:"required"`
		Data       map[string]interface{} `json:"data"`
		DeviceID   string                 `json:"device_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Marshal data to JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid data format", nil)
		return
	}

	event := &SyncEvent{
		Type:       req.Type,
		UserID:     userID,
		ResourceID: req.ResourceID,
		Action:     req.Action,
		Data:       string(dataJSON),
		DeviceID:   req.DeviceID,
		Timestamp:  time.Now(),
	}

	if err := h.service.CreateSyncEvent(c.Request.Context(), event); err != nil {
		h.logger.Error("Failed to create sync event", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to create sync event", nil)
		return
	}

	utils.SuccessResponse(c, event, "Sync event created successfully")
}

// GetOfflineQueue handles GET /api/v1/sync/offline-queue
func (h *Handler) GetOfflineQueue(c *gin.Context) {
	userID := c.GetString("user_id")
	deviceID := c.Query("device_id")

	if deviceID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "device_id is required", nil)
		return
	}

	queue, err := h.service.GetOfflineQueue(c.Request.Context(), userID, deviceID)
	if err != nil {
		h.logger.Error("Failed to get offline queue", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to get offline queue", nil)
		return
	}

	utils.SuccessResponse(c, gin.H{"events": queue}, "Offline queue retrieved successfully")
}

// ProcessOfflineQueue handles POST /api/v1/sync/offline-queue/process
func (h *Handler) ProcessOfflineQueue(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if err := h.service.ProcessOfflineQueue(c.Request.Context(), userID, req.DeviceID); err != nil {
		h.logger.Error("Failed to process offline queue", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to process offline queue", nil)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Offline queue processed successfully"}, "Offline queue processed successfully")
}

// QueueOfflineEvent handles POST /api/v1/sync/offline-queue
func (h *Handler) QueueOfflineEvent(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Type       SyncEventType          `json:"type" binding:"required"`
		ResourceID string                 `json:"resource_id" binding:"required"`
		Action     string                 `json:"action" binding:"required"`
		Data       map[string]interface{} `json:"data"`
		DeviceID   string                 `json:"device_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Marshal data to JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid data format", nil)
		return
	}

	event := &SyncEvent{
		Type:       req.Type,
		UserID:     userID,
		ResourceID: req.ResourceID,
		Action:     req.Action,
		Data:       string(dataJSON),
		DeviceID:   req.DeviceID,
	}

	if err := h.service.QueueOfflineEvent(c.Request.Context(), event); err != nil {
		h.logger.Error("Failed to queue offline event", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to queue offline event", nil)
		return
	}

	utils.SuccessResponse(c, event, "Event queued successfully")
}

// UpdateSyncState handles PUT /api/v1/sync/state
func (h *Handler) UpdateSyncState(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		DeviceID     string `json:"device_id" binding:"required"`
		LastSyncTime int64  `json:"last_sync_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	lastSyncTime := time.Unix(req.LastSyncTime, 0)

	if err := h.service.UpdateSyncState(c.Request.Context(), userID, req.DeviceID, lastSyncTime); err != nil {
		h.logger.Error("Failed to update sync state", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to update sync state", nil)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Sync state updated successfully"}, "Sync state updated successfully")
}
