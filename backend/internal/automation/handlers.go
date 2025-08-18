package automation

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles automation HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new automation handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers automation routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	automation := r.Group("/automation")
	{
		// Webhook endpoints
		webhooks := automation.Group("/webhooks")
		{
			webhooks.POST("", h.CreateWebhookEndpoint)
			webhooks.GET("", h.GetWebhookEndpoints)
			webhooks.PUT("/:id", h.UpdateWebhookEndpoint)
			webhooks.DELETE("/:id", h.DeleteWebhookEndpoint)
			webhooks.GET("/:id/deliveries", h.GetWebhookDeliveries)
		}

		// RSS feeds
		rss := automation.Group("/rss")
		{
			rss.POST("", h.CreateRSSFeed)
			rss.GET("", h.GetRSSFeeds)
			rss.PUT("/:id", h.UpdateRSSFeed)
			rss.DELETE("/:id", h.DeleteRSSFeed)
		}

		// Bulk operations
		bulk := automation.Group("/bulk")
		{
			bulk.POST("", h.CreateBulkOperation)
			bulk.GET("", h.GetBulkOperations)
			bulk.GET("/:id", h.GetBulkOperation)
			bulk.DELETE("/:id", h.CancelBulkOperation)
		}

		// Backup jobs
		backup := automation.Group("/backup")
		{
			backup.POST("", h.CreateBackupJob)
			backup.GET("", h.GetBackupJobs)
			backup.GET("/:id", h.GetBackupJob)
			backup.GET("/:id/download", h.DownloadBackup)
		}

		// API integrations
		integrations := automation.Group("/integrations")
		{
			integrations.POST("", h.CreateAPIIntegration)
			integrations.GET("", h.GetAPIIntegrations)
			integrations.PUT("/:id", h.UpdateAPIIntegration)
			integrations.DELETE("/:id", h.DeleteAPIIntegration)
			integrations.POST("/:id/sync", h.TriggerSync)
			integrations.POST("/:id/test", h.TestIntegration)
		}

		// Automation rules
		rules := automation.Group("/rules")
		{
			rules.POST("", h.CreateAutomationRule)
			rules.GET("", h.GetAutomationRules)
			rules.PUT("/:id", h.UpdateAutomationRule)
			rules.DELETE("/:id", h.DeleteAutomationRule)
			rules.POST("/:id/execute", h.ExecuteAutomationRule)
		}
	}

	// Public RSS feed endpoint
	r.GET("/rss/:publicKey", h.GetPublicRSSFeed)
}

// Webhook Endpoints

// CreateWebhookEndpoint creates a new webhook endpoint
func (h *Handler) CreateWebhookEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req WebhookEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endpoint, err := h.service.CreateWebhookEndpoint(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, endpoint)
}

// GetWebhookEndpoints retrieves webhook endpoints for the authenticated user
func (h *Handler) GetWebhookEndpoints(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	endpoints, err := h.service.GetWebhookEndpoints(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

// UpdateWebhookEndpoint updates a webhook endpoint
func (h *Handler) UpdateWebhookEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	var req WebhookEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endpoint, err := h.service.UpdateWebhookEndpoint(userID, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, endpoint)
}

// DeleteWebhookEndpoint deletes a webhook endpoint
func (h *Handler) DeleteWebhookEndpoint(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	if err := h.service.DeleteWebhookEndpoint(userID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook endpoint deleted successfully"})
}

// GetWebhookDeliveries retrieves webhook deliveries for an endpoint
func (h *Handler) GetWebhookDeliveries(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	deliveries, err := h.service.GetWebhookDeliveries(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deliveries": deliveries})
}

// RSS Feed Endpoints

// CreateRSSFeed creates a new RSS feed
func (h *Handler) CreateRSSFeed(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req RSSFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feed, err := h.service.CreateRSSFeed(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, feed)
}

// GetRSSFeeds retrieves RSS feeds for the authenticated user
func (h *Handler) GetRSSFeeds(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	feeds, err := h.service.GetRSSFeeds(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feeds": feeds})
}

// UpdateRSSFeed updates an RSS feed
func (h *Handler) UpdateRSSFeed(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	var req RSSFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feed, err := h.service.UpdateRSSFeed(userID, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// DeleteRSSFeed deletes an RSS feed
func (h *Handler) DeleteRSSFeed(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	if err := h.service.DeleteRSSFeed(userID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "RSS feed deleted successfully"})
}

// GetPublicRSSFeed serves the public RSS feed
func (h *Handler) GetPublicRSSFeed(c *gin.Context) {
	publicKey := c.Param("publicKey")
	if publicKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public key is required"})
		return
	}

	feed, err := h.service.GetRSSFeedByPublicKey(publicKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "RSS feed not found"})
		return
	}

	content, err := h.service.GenerateRSSContent(feed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate RSS content"})
		return
	}

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.String(http.StatusOK, content)
}

// Bulk Operation Endpoints

// CreateBulkOperation creates a new bulk operation
func (h *Handler) CreateBulkOperation(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req BulkOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operation, err := h.service.CreateBulkOperation(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, operation)
}

// GetBulkOperations retrieves bulk operations for the authenticated user
func (h *Handler) GetBulkOperations(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	operations, err := h.service.GetBulkOperations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"operations": operations})
}

// GetBulkOperation retrieves a specific bulk operation
func (h *Handler) GetBulkOperation(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	operation, err := h.service.GetBulkOperation(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, operation)
}

// CancelBulkOperation cancels a bulk operation
func (h *Handler) CancelBulkOperation(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	if err := h.service.CancelBulkOperation(userID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk operation cancelled successfully"})
}

// Backup Job Endpoints

// CreateBackupJob creates a new backup job
func (h *Handler) CreateBackupJob(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req BackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.service.CreateBackupJob(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// GetBackupJobs retrieves backup jobs for the authenticated user
func (h *Handler) GetBackupJobs(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	jobs, err := h.service.GetBackupJobs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// GetBackupJob retrieves a specific backup job
func (h *Handler) GetBackupJob(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.service.GetBackupJob(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// DownloadBackup downloads a backup file
func (h *Handler) DownloadBackup(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	filePath, err := h.service.GetBackupFilePath(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(filePath)
}

// API Integration Endpoints

// CreateAPIIntegration creates a new API integration
func (h *Handler) CreateAPIIntegration(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req APIIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration, err := h.service.CreateAPIIntegration(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, integration)
}

// GetAPIIntegrations retrieves API integrations for the authenticated user
func (h *Handler) GetAPIIntegrations(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	integrations, err := h.service.GetAPIIntegrations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

// UpdateAPIIntegration updates an API integration
func (h *Handler) UpdateAPIIntegration(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var req APIIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration, err := h.service.UpdateAPIIntegration(userID, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, integration)
}

// DeleteAPIIntegration deletes an API integration
func (h *Handler) DeleteAPIIntegration(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := h.service.DeleteAPIIntegration(userID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API integration deleted successfully"})
}

// TriggerSync triggers a manual sync for an API integration
func (h *Handler) TriggerSync(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	result, err := h.service.TriggerSync(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// TestIntegration tests an API integration
func (h *Handler) TestIntegration(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	result, err := h.service.TestIntegration(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// Automation Rule Endpoints

// CreateAutomationRule creates a new automation rule
func (h *Handler) CreateAutomationRule(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req AutomationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.service.CreateAutomationRule(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetAutomationRules retrieves automation rules for the authenticated user
func (h *Handler) GetAutomationRules(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	rules, err := h.service.GetAutomationRules(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

// UpdateAutomationRule updates an automation rule
func (h *Handler) UpdateAutomationRule(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	var req AutomationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.service.UpdateAutomationRule(userID, uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteAutomationRule deletes an automation rule
func (h *Handler) DeleteAutomationRule(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	if err := h.service.DeleteAutomationRule(userID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Automation rule deleted successfully"})
}

// ExecuteAutomationRule manually executes an automation rule
func (h *Handler) ExecuteAutomationRule(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	result, err := h.service.ExecuteAutomationRule(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}
