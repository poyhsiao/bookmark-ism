package offline

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// OfflineServiceInterface defines the interface for offline service
type OfflineServiceInterface interface {
	CacheBookmark(ctx context.Context, bookmark *database.Bookmark) error
	GetCachedBookmark(ctx context.Context, userID, bookmarkID uint) (*database.Bookmark, error)
	CacheBookmarks(ctx context.Context, bookmarks []database.Bookmark) error
	GetCachedBookmarksForUser(ctx context.Context, userID uint) ([]database.Bookmark, error)
	QueueOfflineChange(ctx context.Context, change *OfflineChange) error
	GetOfflineQueue(ctx context.Context, userID uint) ([]*OfflineChange, error)
	ProcessOfflineQueue(ctx context.Context, userID uint) error
	CheckConnectivity(ctx context.Context) bool
	GetOfflineStatus(ctx context.Context, userID uint) (string, error)
	SetOfflineStatus(ctx context.Context, userID uint, status string) error
	CleanupExpiredCache(ctx context.Context, userID uint) error
	GetCacheStats(ctx context.Context, userID uint) (*CacheStats, error)
	UpdateCacheStats(ctx context.Context, userID uint, stats *CacheStats) error
	ResolveConflict(localChange, serverChange *OfflineChange) *OfflineChange
	SyncWhenOnline(ctx context.Context, userID uint) error
	CreateOfflineChange(userID uint, deviceID, changeType, resourceID, data string) *OfflineChange
	ValidateChangeType(changeType string) bool
	GetOfflineIndicator(ctx context.Context, userID uint) (map[string]interface{}, error)
}

// Handler handles offline-related HTTP requests
type Handler struct {
	service OfflineServiceInterface
}

// NewHandler creates a new offline handler
func NewHandler(service OfflineServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// CacheBookmark caches a bookmark for offline access
func (h *Handler) CacheBookmark(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	var bookmark database.Bookmark
	if err := c.ShouldBindJSON(&bookmark); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", "Invalid bookmark data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Ensure the bookmark belongs to the authenticated user
	bookmark.UserID = uint(userID)

	if err := h.service.CacheBookmark(c.Request.Context(), &bookmark); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CACHE_ERROR", "Failed to cache bookmark", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, "Bookmark cached successfully")
}

// GetCachedBookmark retrieves a cached bookmark
func (h *Handler) GetCachedBookmark(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BOOKMARK_ID", "Invalid bookmark ID", nil)
		return
	}

	bookmark, err := h.service.GetCachedBookmark(c.Request.Context(), uint(userID), uint(bookmarkID))
	if err != nil {
		if err == ErrBookmarkNotCached {
			utils.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Bookmark not found in cache", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CACHE_ERROR", "Failed to get cached bookmark", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, bookmark, "Cached bookmark retrieved successfully")
}

// GetCachedBookmarksForUser retrieves all cached bookmarks for a user
func (h *Handler) GetCachedBookmarksForUser(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	bookmarks, err := h.service.GetCachedBookmarksForUser(c.Request.Context(), uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CACHE_ERROR", "Failed to get cached bookmarks", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, bookmarks, "Cached bookmarks retrieved successfully")
}

// QueueOfflineChange queues a change made while offline
func (h *Handler) QueueOfflineChange(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	var changeData struct {
		DeviceID   string `json:"device_id" binding:"required"`
		Type       string `json:"type" binding:"required"`
		ResourceID string `json:"resource_id" binding:"required"`
		Data       string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&changeData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", "Invalid change data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate change type
	if !h.service.ValidateChangeType(changeData.Type) {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHANGE_TYPE", "Invalid change type", nil)
		return
	}

	// Create offline change
	change := &OfflineChange{
		ID:         strconv.FormatInt(time.Now().UnixNano(), 10),
		UserID:     uint(userID),
		DeviceID:   changeData.DeviceID,
		Type:       changeData.Type,
		ResourceID: changeData.ResourceID,
		Data:       changeData.Data,
		Timestamp:  time.Now(),
		Applied:    false,
	}

	if err := h.service.QueueOfflineChange(c.Request.Context(), change); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "QUEUE_ERROR", "Failed to queue offline change", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, change, "Offline change queued successfully")
}

// GetOfflineQueue retrieves all queued offline changes for a user
func (h *Handler) GetOfflineQueue(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	changes, err := h.service.GetOfflineQueue(c.Request.Context(), uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "QUEUE_ERROR", "Failed to get offline queue", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, changes, "Offline queue retrieved successfully")
}

// ProcessOfflineQueue processes all queued offline changes
func (h *Handler) ProcessOfflineQueue(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	if err := h.service.ProcessOfflineQueue(c.Request.Context(), uint(userID)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "SYNC_ERROR", "Failed to process offline queue", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, "Offline queue processed successfully")
}

// GetOfflineStatus gets the current offline status for a user
func (h *Handler) GetOfflineStatus(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	status, err := h.service.GetOfflineStatus(c.Request.Context(), uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "STATUS_ERROR", "Failed to get offline status", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, map[string]string{"status": status}, "Offline status retrieved successfully")
}

// SetOfflineStatus sets the offline status for a user
func (h *Handler) SetOfflineStatus(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	var statusData struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", "Invalid status data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate status
	if statusData.Status != "online" && statusData.Status != "offline" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_STATUS", "Invalid status value", nil)
		return
	}

	if err := h.service.SetOfflineStatus(c.Request.Context(), uint(userID), statusData.Status); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "STATUS_ERROR", "Failed to set offline status", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, "Offline status updated successfully")
}

// GetCacheStats retrieves cache statistics for a user
func (h *Handler) GetCacheStats(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	stats, err := h.service.GetCacheStats(c.Request.Context(), uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "STATS_ERROR", "Failed to get cache stats", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, stats, "Cache stats retrieved successfully")
}

// CleanupExpiredCache removes expired cache entries for a user
func (h *Handler) CleanupExpiredCache(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	if err := h.service.CleanupExpiredCache(c.Request.Context(), uint(userID)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CLEANUP_ERROR", "Failed to cleanup expired cache", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, "Expired cache cleaned up successfully")
}

// GetOfflineIndicator retrieves offline indicator information
func (h *Handler) GetOfflineIndicator(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	indicator, err := h.service.GetOfflineIndicator(c.Request.Context(), uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INDICATOR_ERROR", "Failed to get offline indicator", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, indicator, "Offline indicator retrieved successfully")
}

// CheckConnectivity checks if the service is online
func (h *Handler) CheckConnectivity(c *gin.Context) {
	isOnline := h.service.CheckConnectivity(c.Request.Context())

	utils.SuccessResponse(c, map[string]bool{"is_online": isOnline}, "Connectivity check completed")
}
