package storage

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for storage operations
type Handler struct {
	service *Service
}

// NewHandler creates a new storage handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// UploadScreenshotRequest represents the request for uploading a screenshot
type UploadScreenshotRequest struct {
	BookmarkID string `json:"bookmark_id" binding:"required"`
}

// UploadScreenshotResponse represents the response for uploading a screenshot
type UploadScreenshotResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// UploadScreenshot handles screenshot upload
func (h *Handler) UploadScreenshot(c *gin.Context) {
	var req UploadScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get file from form data
	file, header, err := c.Request.FormFile("screenshot")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "FILE_REQUIRED", "No screenshot file provided", map[string]interface{}{"error": err.Error()})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "File must be an image", nil)
		return
	}

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FILE_READ_ERROR", "Failed to read file", map[string]interface{}{"error": err.Error()})
		return
	}

	// Store screenshot
	url, err := h.service.StoreScreenshot(c.Request.Context(), req.BookmarkID, data)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store screenshot", map[string]interface{}{"error": err.Error()})
		return
	}

	response := UploadScreenshotResponse{
		URL: url,
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Screenshot uploaded successfully",
		Data:    response,
	})
}

// UploadAvatarRequest represents the request for uploading an avatar
type UploadAvatarRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// UploadAvatarResponse represents the response for uploading an avatar
type UploadAvatarResponse struct {
	URL string `json:"url"`
}

// UploadAvatar handles avatar upload
func (h *Handler) UploadAvatar(c *gin.Context) {
	var req UploadAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get file from form data
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "FILE_REQUIRED", "No avatar file provided", map[string]interface{}{"error": err.Error()})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "File must be an image", nil)
		return
	}

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FILE_READ_ERROR", "Failed to read file", map[string]interface{}{"error": err.Error()})
		return
	}

	// Store avatar
	url, err := h.service.StoreAvatar(c.Request.Context(), req.UserID, data, contentType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store avatar", map[string]interface{}{"error": err.Error()})
		return
	}

	response := UploadAvatarResponse{
		URL: url,
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Avatar uploaded successfully",
		Data:    response,
	})
}

// GetFileURLRequest represents the request for getting a file URL
type GetFileURLRequest struct {
	ObjectName string `json:"object_name" binding:"required"`
	ExpiryHour int    `json:"expiry_hour,omitempty"`
}

// GetFileURLResponse represents the response for getting a file URL
type GetFileURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetFileURL handles getting a presigned URL for a file
func (h *Handler) GetFileURL(c *gin.Context) {
	var req GetFileURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Default expiry to 1 hour
	expiryHour := req.ExpiryHour
	if expiryHour == 0 {
		expiryHour = 1
	}

	expiry := time.Duration(expiryHour) * time.Hour
	url, err := h.service.GetFileURL(c.Request.Context(), req.ObjectName, expiry)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "URL_GENERATION_ERROR", "Failed to get file URL", map[string]interface{}{"error": err.Error()})
		return
	}

	response := GetFileURLResponse{
		URL:       url,
		ExpiresAt: time.Now().Add(expiry),
	}

	utils.SuccessResponse(c, response, "File URL generated successfully")
}

// DeleteFileRequest represents the request for deleting a file
type DeleteFileRequest struct {
	ObjectName string `json:"object_name" binding:"required"`
}

// DeleteFile handles file deletion
func (h *Handler) DeleteFile(c *gin.Context) {
	var req DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	err := h.service.DeleteFile(c.Request.Context(), req.ObjectName)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete file", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, "File deleted successfully")
}

// HealthCheck handles storage health check
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "HEALTH_CHECK_FAILED", "Storage service unhealthy", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, "Storage service healthy")
}

// ServeFile handles serving files directly from storage
func (h *Handler) ServeFile(c *gin.Context) {
	objectName := c.Param("path")
	if objectName == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "PATH_REQUIRED", "File path is required", nil)
		return
	}

	// Get expiry from query parameter (default 1 hour)
	expiryStr := c.DefaultQuery("expiry", "1")
	expiryHour, err := strconv.Atoi(expiryStr)
	if err != nil {
		expiryHour = 1
	}

	expiry := time.Duration(expiryHour) * time.Hour
	url, err := h.service.GetFileURL(c.Request.Context(), objectName, expiry)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "FILE_NOT_FOUND", "File not found", map[string]interface{}{"error": err.Error()})
		return
	}

	// Redirect to the presigned URL
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// RegisterRoutes registers storage routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	storage := router.Group("/storage")
	{
		storage.POST("/screenshot", h.UploadScreenshot)
		storage.POST("/avatar", h.UploadAvatar)
		storage.POST("/file-url", h.GetFileURL)
		storage.DELETE("/file", h.DeleteFile)
		storage.GET("/health", h.HealthCheck)
		storage.GET("/file/*path", h.ServeFile)
	}
}
