package monitoring

import (
	"net/http"
	"strconv"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for monitoring operations
type Handler struct {
	service *Service
}

// NewHandler creates a new monitoring handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers monitoring routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	monitoring := router.Group("/monitoring")
	{
		// Link checking endpoints
		monitoring.POST("/check-link", h.CheckLink)
		monitoring.GET("/bookmarks/:bookmark_id/checks", h.GetLinkChecks)

		// Monitoring job endpoints
		monitoring.POST("/jobs", h.CreateMonitoringJob)
		monitoring.GET("/jobs", h.ListMonitoringJobs)
		monitoring.GET("/jobs/:job_id", h.GetMonitoringJob)
		monitoring.PUT("/jobs/:job_id", h.UpdateMonitoringJob)
		monitoring.DELETE("/jobs/:job_id", h.DeleteMonitoringJob)

		// Maintenance report endpoints
		monitoring.POST("/reports", h.GenerateMaintenanceReport)

		// Notification endpoints
		monitoring.GET("/notifications", h.GetNotifications)
		monitoring.PUT("/notifications/:notification_id/read", h.MarkNotificationAsRead)
	}
}

// CheckLink handles link checking requests
func (h *Handler) CheckLink(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var req CreateLinkCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	linkCheck, err := h.service.CheckLink(c.Request.Context(), userID, &req)
	if err != nil {
		if err.Error() == "bookmark not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "BOOKMARK_NOT_FOUND", "Bookmark not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CHECK_FAILED", "Failed to check link", map[string]interface{}{"error": err.Error()})
		return
	}

	response := LinkCheckResponse{
		LinkCheck: linkCheck,
		Message:   "Link check completed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetLinkChecks handles requests to get link checks for a bookmark
func (h *Handler) GetLinkChecks(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	bookmarkIDStr := c.Param("bookmark_id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_BOOKMARK_ID", "Invalid bookmark ID", nil)
		return
	}

	page, pageSize := utils.GetPaginationParams(c)

	checks, total, err := h.service.GetLinkChecks(c.Request.Context(), userID, uint(bookmarkID), page, pageSize)
	if err != nil {
		if err.Error() == "bookmark not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "BOOKMARK_NOT_FOUND", "Bookmark not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get link checks", map[string]interface{}{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	response := ListResponse{
		Items:      checks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// CreateMonitoringJob handles monitoring job creation requests
func (h *Handler) CreateMonitoringJob(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var req CreateMonitoringJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	job, err := h.service.CreateMonitoringJob(c.Request.Context(), userID, &req)
	if err != nil {
		if err.Error() == "invalid cron expression" {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CRON", "Invalid cron expression", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create monitoring job", map[string]interface{}{"error": err.Error()})
		return
	}

	response := MonitoringJobResponse{
		Job:     job,
		Message: "Monitoring job created successfully",
	}

	c.JSON(http.StatusCreated, response)
}

// ListMonitoringJobs handles requests to list monitoring jobs
func (h *Handler) ListMonitoringJobs(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	page, pageSize := utils.GetPaginationParams(c)

	jobs, total, err := h.service.ListMonitoringJobs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to list monitoring jobs", map[string]interface{}{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	response := ListResponse{
		Items:      jobs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetMonitoringJob handles requests to get a specific monitoring job
func (h *Handler) GetMonitoringJob(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	jobIDStr := c.Param("job_id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_JOB_ID", "Invalid job ID", nil)
		return
	}

	job, err := h.service.GetMonitoringJob(c.Request.Context(), userID, uint(jobID))
	if err != nil {
		if err.Error() == "monitoring job not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "JOB_NOT_FOUND", "Monitoring job not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get monitoring job", map[string]interface{}{"error": err.Error()})
		return
	}

	response := MonitoringJobResponse{
		Job:     job,
		Message: "Monitoring job retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// UpdateMonitoringJob handles monitoring job update requests
func (h *Handler) UpdateMonitoringJob(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	jobIDStr := c.Param("job_id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_JOB_ID", "Invalid job ID", nil)
		return
	}

	var req UpdateMonitoringJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", map[string]interface{}{"error": err.Error()})
		return
	}

	job, err := h.service.UpdateMonitoringJob(c.Request.Context(), userID, uint(jobID), &req)
	if err != nil {
		if err.Error() == "monitoring job not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "JOB_NOT_FOUND", "Monitoring job not found", nil)
			return
		}
		if err.Error() == "invalid cron expression" {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CRON", "Invalid cron expression", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update monitoring job", map[string]interface{}{"error": err.Error()})
		return
	}

	response := MonitoringJobResponse{
		Job:     job,
		Message: "Monitoring job updated successfully",
	}

	c.JSON(http.StatusOK, response)
}

// DeleteMonitoringJob handles monitoring job deletion requests
func (h *Handler) DeleteMonitoringJob(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	jobIDStr := c.Param("job_id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_JOB_ID", "Invalid job ID", nil)
		return
	}

	err = h.service.DeleteMonitoringJob(c.Request.Context(), userID, uint(jobID))
	if err != nil {
		if err.Error() == "monitoring job not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "JOB_NOT_FOUND", "Monitoring job not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete monitoring job", map[string]interface{}{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Monitoring job deleted successfully",
	})
}

// GenerateMaintenanceReport handles maintenance report generation requests
func (h *Handler) GenerateMaintenanceReport(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	// Optional collection ID filter
	var collectionID *uint
	if collectionIDStr := c.Query("collection_id"); collectionIDStr != "" {
		if id, err := strconv.ParseUint(collectionIDStr, 10, 32); err == nil {
			collectionIDUint := uint(id)
			collectionID = &collectionIDUint
		}
	}

	report, err := h.service.GenerateMaintenanceReport(c.Request.Context(), userID, collectionID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "REPORT_FAILED", "Failed to generate maintenance report", map[string]interface{}{"error": err.Error()})
		return
	}

	response := MaintenanceReportResponse{
		Report:  report,
		Message: "Maintenance report generated successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetNotifications handles requests to get user notifications
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	page, pageSize := utils.GetPaginationParams(c)
	unreadOnly := c.Query("unread_only") == "true"

	notifications, total, err := h.service.GetNotifications(c.Request.Context(), userID, page, pageSize, unreadOnly)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to get notifications", map[string]interface{}{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	response := ListResponse{
		Items:      notifications,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// MarkNotificationAsRead handles requests to mark a notification as read
func (h *Handler) MarkNotificationAsRead(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_NOTIFICATION_ID", "Invalid notification ID", nil)
		return
	}

	err = h.service.MarkNotificationAsRead(c.Request.Context(), userID, uint(notificationID))
	if err != nil {
		if err.Error() == "notification not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Notification not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to mark notification as read", map[string]interface{}{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read successfully",
	})
}
