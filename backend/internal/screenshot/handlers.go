package screenshot

import (
	"context"
	"net/http"
	"strconv"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ScreenshotServiceInterface defines the interface for screenshot service
type ScreenshotServiceInterface interface {
	CaptureScreenshot(ctx context.Context, bookmarkID, pageURL string, opts CaptureOptions) (*CaptureResult, error)
	UpdateBookmarkScreenshot(ctx context.Context, bookmarkID, pageURL string) (*CaptureResult, error)
	GetFavicon(ctx context.Context, pageURL string) ([]byte, error)
	CaptureFromURL(ctx context.Context, pageURL string) ([]byte, error)
}

// Handler handles HTTP requests for screenshot operations
type Handler struct {
	service ScreenshotServiceInterface
}

// NewHandler creates a new screenshot handler
func NewHandler(service ScreenshotServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// CaptureScreenshotRequest represents the request for capturing a screenshot
type CaptureScreenshotRequest struct {
	BookmarkID string `json:"bookmark_id" binding:"required"`
	URL        string `json:"url" binding:"required"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Quality    int    `json:"quality,omitempty"`
	Format     string `json:"format,omitempty"`
	Thumbnail  bool   `json:"thumbnail,omitempty"`
}

// CaptureScreenshotResponse represents the response for capturing a screenshot
type CaptureScreenshotResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Size         int64  `json:"size"`
	Format       string `json:"format"`
}

// CaptureScreenshot handles screenshot capture requests
func (h *Handler) CaptureScreenshot(c *gin.Context) {
	var req CaptureScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set default options
	opts := CaptureOptions{
		Width:     req.Width,
		Height:    req.Height,
		Quality:   req.Quality,
		Format:    req.Format,
		Thumbnail: req.Thumbnail,
	}

	// Apply defaults
	if opts.Width == 0 {
		opts.Width = 1200
	}
	if opts.Height == 0 {
		opts.Height = 800
	}
	if opts.Quality == 0 {
		opts.Quality = 85
	}
	if opts.Format == "" {
		opts.Format = "jpeg"
	}

	// Capture screenshot
	result, err := h.service.CaptureScreenshot(c.Request.Context(), req.BookmarkID, req.URL, opts)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CAPTURE_ERROR", "Failed to capture screenshot", map[string]interface{}{"error": err.Error()})
		return
	}

	response := CaptureScreenshotResponse{
		URL:          result.URL,
		ThumbnailURL: result.ThumbnailURL,
		Width:        result.Width,
		Height:       result.Height,
		Size:         result.Size,
		Format:       result.Format,
	}

	utils.SuccessResponse(c, response, "Screenshot captured successfully")
}

// UpdateBookmarkScreenshotRequest represents the request for updating a bookmark screenshot
type UpdateBookmarkScreenshotRequest struct {
	URL string `json:"url" binding:"required"`
}

// UpdateBookmarkScreenshot handles bookmark screenshot update requests
func (h *Handler) UpdateBookmarkScreenshot(c *gin.Context) {
	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_PARAMETER", "Bookmark ID is required", nil)
		return
	}

	var req UpdateBookmarkScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Update screenshot
	result, err := h.service.UpdateBookmarkScreenshot(c.Request.Context(), bookmarkID, req.URL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update screenshot", map[string]interface{}{"error": err.Error()})
		return
	}

	response := CaptureScreenshotResponse{
		URL:          result.URL,
		ThumbnailURL: result.ThumbnailURL,
		Width:        result.Width,
		Height:       result.Height,
		Size:         result.Size,
		Format:       result.Format,
	}

	utils.SuccessResponse(c, response, "Screenshot updated successfully")
}

// GetFaviconRequest represents the request for getting a favicon
type GetFaviconRequest struct {
	URL string `json:"url" binding:"required"`
}

// GetFaviconResponse represents the response for getting a favicon
type GetFaviconResponse struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
}

// GetFavicon handles favicon retrieval requests
func (h *Handler) GetFavicon(c *gin.Context) {
	var req GetFaviconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get favicon
	faviconData, err := h.service.GetFavicon(c.Request.Context(), req.URL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "FAVICON_NOT_FOUND", "Favicon not found", map[string]interface{}{"error": err.Error()})
		return
	}

	// Return favicon data directly
	c.Header("Content-Type", "image/x-icon")
	c.Header("Content-Length", strconv.Itoa(len(faviconData)))
	c.Data(http.StatusOK, "image/x-icon", faviconData)
}

// CaptureFromURLRequest represents the request for capturing from URL
type CaptureFromURLRequest struct {
	URL string `json:"url" binding:"required"`
}

// CaptureFromURL handles direct URL screenshot capture
func (h *Handler) CaptureFromURL(c *gin.Context) {
	var req CaptureFromURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", map[string]interface{}{"error": err.Error()})
		return
	}

	// Capture screenshot
	screenshotData, err := h.service.CaptureFromURL(c.Request.Context(), req.URL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "CAPTURE_ERROR", "Failed to capture screenshot", map[string]interface{}{"error": err.Error()})
		return
	}

	// Return screenshot data directly
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", strconv.Itoa(len(screenshotData)))
	c.Data(http.StatusOK, "image/png", screenshotData)
}

// RegisterRoutes registers screenshot routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	screenshot := router.Group("/screenshot")
	{
		screenshot.POST("/capture", h.CaptureScreenshot)
		screenshot.PUT("/bookmark/:id", h.UpdateBookmarkScreenshot)
		screenshot.POST("/favicon", h.GetFavicon)
		screenshot.POST("/url", h.CaptureFromURL)
	}
}
