package import_export

import (
	"fmt"
	"net/http"
	"strings"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers provides HTTP handlers for import/export functionality
type Handlers struct {
	service *Service
}

// NewHandlers creates new import/export handlers
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRoutes registers import/export routes
func (h *Handlers) RegisterRoutes(router *gin.RouterGroup) {
	importExport := router.Group("/import-export")
	{
		// Import endpoints
		importExport.POST("/import/chrome", h.ImportFromChrome)
		importExport.POST("/import/firefox", h.ImportFromFirefox)
		importExport.POST("/import/safari", h.ImportFromSafari)
		importExport.GET("/import/progress/:jobId", h.GetImportProgress)

		// Export endpoints
		importExport.GET("/export/json", h.ExportToJSON)
		importExport.GET("/export/html", h.ExportToHTML)

		// Utility endpoints
		importExport.POST("/detect-duplicates", h.DetectDuplicates)
	}
}

// ImportFromChrome handles Chrome bookmark import
func (h *Handlers) ImportFromChrome(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE", "No file provided or invalid file", map[string]interface{}{"error": err.Error()})
		return
	}
	defer file.Close()

	// Validate file extension (more reliable than content type in multipart uploads)
	if !strings.HasSuffix(header.Filename, ".json") {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "File must be a JSON file", nil)
		return
	}

	// Generate job ID for progress tracking
	jobID := uuid.New().String()

	// Start import process
	result, err := h.service.ImportBookmarksFromChrome(c.Request.Context(), userID.(uint), file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "IMPORT_FAILED", "Failed to import Chrome bookmarks", map[string]interface{}{"error": err.Error()})
		return
	}

	// Return result with job ID
	response := gin.H{
		"job_id": jobID,
		"result": result,
	}

	utils.SuccessResponse(c, response, "Chrome bookmarks imported successfully")
}

// ImportFromFirefox handles Firefox bookmark import
func (h *Handlers) ImportFromFirefox(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE", "No file provided or invalid file", map[string]interface{}{"error": err.Error()})
		return
	}
	defer file.Close()

	// Validate file extension (more reliable than content type in multipart uploads)
	if !isHTMLFile(header.Filename) {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "File must be an HTML file", nil)
		return
	}

	// Generate job ID for progress tracking
	jobID := uuid.New().String()

	// Start import process
	result, err := h.service.ImportBookmarksFromFirefox(c.Request.Context(), userID.(uint), file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "IMPORT_FAILED", "Failed to import Firefox bookmarks", map[string]interface{}{"error": err.Error()})
		return
	}

	// Return result with job ID
	response := gin.H{
		"job_id": jobID,
		"result": result,
	}

	utils.SuccessResponse(c, response, "Firefox bookmarks imported successfully")
}

// ImportFromSafari handles Safari bookmark import
func (h *Handlers) ImportFromSafari(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE", "No file provided or invalid file", map[string]interface{}{"error": err.Error()})
		return
	}
	defer file.Close()

	// Validate file type
	if !isPlistFile(header.Filename) {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "File must be a plist file", nil)
		return
	}

	// Generate job ID for progress tracking
	jobID := uuid.New().String()

	// Start import process
	result, err := h.service.ImportBookmarksFromSafari(c.Request.Context(), userID.(uint), file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "IMPORT_FAILED", "Failed to import Safari bookmarks", map[string]interface{}{"error": err.Error()})
		return
	}

	// Return result with job ID
	response := gin.H{
		"job_id": jobID,
		"result": result,
	}

	utils.SuccessResponse(c, response, "Safari bookmarks imported successfully")
}

// GetImportProgress handles import progress requests
func (h *Handlers) GetImportProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	jobID := c.Param("jobId")
	if jobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "Job ID is required", nil)
		return
	}

	progress, err := h.service.GetImportProgress(c.Request.Context(), userID.(uint), jobID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "PROGRESS_FAILED", "Failed to get import progress", map[string]interface{}{"error": err.Error()})
		return
	}

	if progress == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "JOB_NOT_FOUND", "Import job not found", nil)
		return
	}

	utils.SuccessResponse(c, progress, "Import progress retrieved successfully")
}

// ExportToJSON handles JSON export
func (h *Handlers) ExportToJSON(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Set response headers for file download
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=bookmarks_%d.json", userID))

	// Export bookmarks
	if err := h.service.ExportBookmarksToJSON(c.Request.Context(), userID.(uint), c.Writer); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to export bookmarks to JSON", map[string]interface{}{"error": err.Error()})
		return
	}
}

// ExportToHTML handles HTML export
func (h *Handlers) ExportToHTML(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Set response headers for file download
	c.Header("Content-Type", "text/html")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=bookmarks_%d.html", userID))

	// Export bookmarks
	if err := h.service.ExportBookmarksToHTML(c.Request.Context(), userID.(uint), c.Writer); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to export bookmarks to HTML", map[string]interface{}{"error": err.Error()})
		return
	}
}

// DetectDuplicates handles duplicate detection requests
func (h *Handlers) DetectDuplicates(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var request struct {
		URLs []string `json:"urls" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if len(request.URLs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "At least one URL is required", nil)
		return
	}

	if len(request.URLs) > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "TOO_MANY_URLS", "Maximum 100 URLs allowed per request", nil)
		return
	}

	// Check each URL for duplicates
	duplicates := make(map[string]bool)
	for _, url := range request.URLs {
		isDuplicate, err := h.service.DetectDuplicate(c.Request.Context(), userID.(uint), url)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "DUPLICATE_CHECK_FAILED", "Failed to check for duplicates", map[string]interface{}{"error": err.Error()})
			return
		}
		duplicates[url] = isDuplicate
	}

	response := gin.H{
		"duplicates": duplicates,
		"total_urls": len(request.URLs),
		"duplicate_count": func() int {
			count := 0
			for _, isDup := range duplicates {
				if isDup {
					count++
				}
			}
			return count
		}(),
	}

	utils.SuccessResponse(c, response, "Duplicate check completed successfully")
}

// Helper functions

// isHTMLFile checks if the filename has HTML extension
func isHTMLFile(filename string) bool {
	return len(filename) > 5 && (filename[len(filename)-5:] == ".html" || filename[len(filename)-4:] == ".htm")
}

// isPlistFile checks if the filename has plist extension
func isPlistFile(filename string) bool {
	return len(filename) > 6 && filename[len(filename)-6:] == ".plist"
}
