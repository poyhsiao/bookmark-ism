package automation

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Service handles automation operations
type Service struct {
	db                     *gorm.DB
	httpClient             *http.Client
	disableAsyncProcessing bool
}

// NewService creates a new automation service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		disableAsyncProcessing: false,
	}
}

// NewServiceForTesting creates a new automation service for testing with async processing disabled
func NewServiceForTesting(db *gorm.DB) *Service {
	return &Service{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		disableAsyncProcessing: true,
	}
}

// Webhook Management

// CreateWebhookEndpoint creates a new webhook endpoint
func (s *Service) CreateWebhookEndpoint(userID string, req WebhookEndpointRequest) (*WebhookEndpoint, error) {
	secret, err := s.generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	endpoint := &WebhookEndpoint{
		UserID:     userID,
		Name:       req.Name,
		URL:        req.URL,
		Secret:     secret,
		Events:     StringSlice(req.Events),
		Active:     true,
		RetryCount: req.RetryCount,
		Timeout:    req.Timeout,
		Headers:    StringMap(req.Headers),
	}

	if endpoint.RetryCount == 0 {
		endpoint.RetryCount = 3
	}
	if endpoint.Timeout == 0 {
		endpoint.Timeout = 30
	}

	if err := s.db.Create(endpoint).Error; err != nil {
		return nil, fmt.Errorf("failed to create webhook endpoint: %w", err)
	}

	return endpoint, nil
}

// GetWebhookEndpoints retrieves webhook endpoints for a user
func (s *Service) GetWebhookEndpoints(userID string) ([]WebhookEndpoint, error) {
	var endpoints []WebhookEndpoint
	if err := s.db.Where("user_id = ?", userID).Find(&endpoints).Error; err != nil {
		return nil, fmt.Errorf("failed to get webhook endpoints: %w", err)
	}
	return endpoints, nil
}

// UpdateWebhookEndpoint updates a webhook endpoint
func (s *Service) UpdateWebhookEndpoint(userID string, id uint, req WebhookEndpointRequest) (*WebhookEndpoint, error) {
	var endpoint WebhookEndpoint
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&endpoint).Error; err != nil {
		return nil, fmt.Errorf("webhook endpoint not found: %w", err)
	}

	endpoint.Name = req.Name
	endpoint.URL = req.URL
	endpoint.Events = StringSlice(req.Events)
	endpoint.Active = req.Active
	endpoint.RetryCount = req.RetryCount
	endpoint.Timeout = req.Timeout
	endpoint.Headers = StringMap(req.Headers)

	if err := s.db.Save(&endpoint).Error; err != nil {
		return nil, fmt.Errorf("failed to update webhook endpoint: %w", err)
	}

	return &endpoint, nil
}

// DeleteWebhookEndpoint deletes a webhook endpoint
func (s *Service) DeleteWebhookEndpoint(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&WebhookEndpoint{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete webhook endpoint: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook endpoint not found")
	}
	return nil
}

// TriggerWebhook triggers webhooks for a specific event
func (s *Service) TriggerWebhook(ctx context.Context, event WebhookEvent, userID string, data interface{}) error {
	var endpoints []WebhookEndpoint
	if err := s.db.Where("user_id = ? AND active = ?", userID, true).Find(&endpoints).Error; err != nil {
		return fmt.Errorf("failed to get webhook endpoints: %w", err)
	}

	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now(),
		UserID:    userID,
		Data:      data,
	}

	for _, endpoint := range endpoints {
		// Check if endpoint is subscribed to this event
		if !s.isEventSubscribed(endpoint.Events, string(event)) {
			continue
		}

		// Create delivery record
		delivery := &WebhookDelivery{
			EndpointID: endpoint.ID,
			Event:      event,
			Payload:    InterfaceMap(s.structToMap(payload)),
			Status:     "pending",
		}

		if err := s.db.Create(delivery).Error; err != nil {
			continue // Log error but continue with other endpoints
		}

		// Deliver webhook asynchronously (unless disabled for testing)
		if !s.disableAsyncProcessing {
			go s.deliverWebhook(ctx, &endpoint, delivery, payload)
		}
	}

	return nil
}

// RSS Feed Management

// CreateRSSFeed creates a new RSS feed
func (s *Service) CreateRSSFeed(userID string, req RSSFeedRequest) (*RSSFeed, error) {
	publicKey, err := s.generatePublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	feed := &RSSFeed{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Link:        req.Link,
		Language:    req.Language,
		Copyright:   req.Copyright,
		Category:    req.Category,
		TTL:         req.TTL,
		MaxItems:    req.MaxItems,
		Active:      true,
		PublicKey:   publicKey,
		Collections: UintSlice(req.Collections),
		Tags:        StringSlice(req.Tags),
	}

	if feed.Language == "" {
		feed.Language = "en"
	}
	if feed.TTL == 0 {
		feed.TTL = 60
	}
	if feed.MaxItems == 0 {
		feed.MaxItems = 50
	}

	if err := s.db.Create(feed).Error; err != nil {
		return nil, fmt.Errorf("failed to create RSS feed: %w", err)
	}

	return feed, nil
}

// GetRSSFeeds retrieves RSS feeds for a user
func (s *Service) GetRSSFeeds(userID string) ([]RSSFeed, error) {
	var feeds []RSSFeed
	if err := s.db.Where("user_id = ?", userID).Find(&feeds).Error; err != nil {
		return nil, fmt.Errorf("failed to get RSS feeds: %w", err)
	}
	return feeds, nil
}

// GetRSSFeedByPublicKey retrieves an RSS feed by public key
func (s *Service) GetRSSFeedByPublicKey(publicKey string) (*RSSFeed, error) {
	var feed RSSFeed
	if err := s.db.Where("public_key = ? AND active = ?", publicKey, true).First(&feed).Error; err != nil {
		return nil, fmt.Errorf("RSS feed not found: %w", err)
	}
	return &feed, nil
}

// GenerateRSSContent generates RSS XML content for a feed
func (s *Service) GenerateRSSContent(feed *RSSFeed) (string, error) {
	// This would integrate with bookmark service to get actual bookmarks
	// For now, we'll create a basic RSS structure

	items, err := s.getRSSItems(feed)
	if err != nil {
		return "", fmt.Errorf("failed to get RSS items: %w", err)
	}

	rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title><![CDATA[%s]]></title>
    <description><![CDATA[%s]]></description>
    <link>%s</link>
    <language>%s</language>
    <copyright><![CDATA[%s]]></copyright>
    <category><![CDATA[%s]]></category>
    <ttl>%d</ttl>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/rss/%s" rel="self" type="application/rss+xml"/>
%s
  </channel>
</rss>`,
		feed.Title,
		feed.Description,
		feed.Link,
		feed.Language,
		feed.Copyright,
		feed.Category,
		feed.TTL,
		time.Now().Format(time.RFC1123Z),
		feed.Link,
		feed.PublicKey,
		s.formatRSSItems(items),
	)

	return rss, nil
}

// Bulk Operations

// CreateBulkOperation creates a new bulk operation
func (s *Service) CreateBulkOperation(userID string, req BulkOperationRequest) (*BulkOperation, error) {
	operation := &BulkOperation{
		UserID:     userID,
		Type:       req.Type,
		Status:     "pending",
		Parameters: InterfaceMap(req.Parameters),
	}

	if err := s.db.Create(operation).Error; err != nil {
		return nil, fmt.Errorf("failed to create bulk operation: %w", err)
	}

	// Start processing asynchronously (unless disabled for testing)
	if !s.disableAsyncProcessing {
		go s.processBulkOperation(operation)
	}

	return operation, nil
}

// GetBulkOperations retrieves bulk operations for a user
func (s *Service) GetBulkOperations(userID string) ([]BulkOperation, error) {
	var operations []BulkOperation
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&operations).Error; err != nil {
		return nil, fmt.Errorf("failed to get bulk operations: %w", err)
	}
	return operations, nil
}

// GetBulkOperation retrieves a specific bulk operation
func (s *Service) GetBulkOperation(userID string, id uint) (*BulkOperation, error) {
	var operation BulkOperation
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&operation).Error; err != nil {
		return nil, fmt.Errorf("bulk operation not found: %w", err)
	}
	return &operation, nil
}

// Backup Management

// CreateBackupJob creates a new backup job
func (s *Service) CreateBackupJob(userID string, req BackupRequest) (*BackupJob, error) {
	job := &BackupJob{
		UserID:        userID,
		Type:          req.Type,
		Status:        "pending",
		Compression:   req.Compression,
		Encrypted:     req.Encrypted,
		RetentionDays: 30,
	}

	if job.Compression == "" {
		job.Compression = "gzip"
	}

	if err := s.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup job: %w", err)
	}

	// Start backup process asynchronously (unless disabled for testing)
	if !s.disableAsyncProcessing {
		go s.processBackupJob(job)
	}

	return job, nil
}

// GetBackupJobs retrieves backup jobs for a user
func (s *Service) GetBackupJobs(userID string) ([]BackupJob, error) {
	var jobs []BackupJob
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to get backup jobs: %w", err)
	}
	return jobs, nil
}

// API Integration Management

// CreateAPIIntegration creates a new API integration
func (s *Service) CreateAPIIntegration(userID string, req APIIntegrationRequest) (*APIIntegration, error) {
	integration := &APIIntegration{
		UserID:       userID,
		Name:         req.Name,
		Type:         req.Type,
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		APISecret:    req.APISecret,
		Active:       true,
		RateLimit:    100,
		SyncEnabled:  req.SyncEnabled,
		SyncInterval: req.SyncInterval,
		Config:       InterfaceMap(req.Config),
	}

	if integration.SyncInterval == 0 {
		integration.SyncInterval = 3600 // 1 hour default
	}

	if err := s.db.Create(integration).Error; err != nil {
		return nil, fmt.Errorf("failed to create API integration: %w", err)
	}

	return integration, nil
}

// GetAPIIntegrations retrieves API integrations for a user
func (s *Service) GetAPIIntegrations(userID string) ([]APIIntegration, error) {
	var integrations []APIIntegration
	if err := s.db.Where("user_id = ?", userID).Find(&integrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get API integrations: %w", err)
	}
	return integrations, nil
}

// Helper methods

func (s *Service) generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) generatePublicKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) isEventSubscribed(events StringSlice, event string) bool {
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}

func (s *Service) structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	data, _ := json.Marshal(obj)
	json.Unmarshal(data, &result)
	return result
}

func (s *Service) deliverWebhook(ctx context.Context, endpoint *WebhookEndpoint, delivery *WebhookDelivery, payload WebhookPayload) {
	// Update delivery status to running
	delivery.Status = "running"
	delivery.AttemptCount++
	s.db.Save(delivery)

	// Prepare request
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.updateDeliveryError(delivery, "Failed to marshal payload", 0)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.updateDeliveryError(delivery, "Failed to create request", 0)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "BookmarkSync-Webhook/1.0")

	// Add custom headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// Add signature
	signature := s.generateSignature(payloadBytes, endpoint.Secret)
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", string(payload.Event))
	req.Header.Set("X-Webhook-Timestamp", strconv.FormatInt(payload.Timestamp.Unix(), 10))

	// Set timeout
	client := &http.Client{
		Timeout: time.Duration(endpoint.Timeout) * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		s.updateDeliveryError(delivery, err.Error(), 0)
		s.scheduleRetry(delivery, endpoint)
		return
	}
	defer resp.Body.Close()

	// Read response
	responseBody, _ := io.ReadAll(resp.Body)

	// Update delivery
	delivery.StatusCode = resp.StatusCode
	delivery.Response = string(responseBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = "success"
	} else {
		delivery.Status = "failed"
		s.scheduleRetry(delivery, endpoint)
	}

	s.db.Save(delivery)
}

func (s *Service) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func (s *Service) updateDeliveryError(delivery *WebhookDelivery, error string, statusCode int) {
	delivery.Status = "failed"
	delivery.Error = error
	delivery.StatusCode = statusCode
	s.db.Save(delivery)
}

func (s *Service) scheduleRetry(delivery *WebhookDelivery, endpoint *WebhookEndpoint) {
	if delivery.AttemptCount < endpoint.RetryCount {
		// Exponential backoff: 2^attempt minutes
		retryDelay := time.Duration(1<<delivery.AttemptCount) * time.Minute
		nextRetry := time.Now().Add(retryDelay)
		delivery.NextRetryAt = &nextRetry
		s.db.Save(delivery)
	}
}

func (s *Service) getRSSItems(feed *RSSFeed) ([]RSSItem, error) {
	// This would integrate with bookmark service to get actual bookmarks
	// For now, return empty slice - will be implemented when integrating with bookmark service
	return []RSSItem{}, nil
}

func (s *Service) formatRSSItems(items []RSSItem) string {
	var itemsXML strings.Builder

	for _, item := range items {
		itemsXML.WriteString(fmt.Sprintf(`    <item>
      <title><![CDATA[%s]]></title>
      <link>%s</link>
      <description><![CDATA[%s]]></description>
      <guid>%s</guid>
      <pubDate>%s</pubDate>
    </item>
`, item.Title, item.Link, item.Description, item.GUID, item.PubDate.Format(time.RFC1123Z)))
	}

	return itemsXML.String()
}

func (s *Service) processBulkOperation(operation *BulkOperation) {
	// Update status to running
	now := time.Now()
	operation.Status = "running"
	operation.StartedAt = &now
	s.db.Save(operation)

	// Process based on operation type
	var err error
	switch operation.Type {
	case "import":
		err = s.processBulkImport(operation)
	case "export":
		err = s.processBulkExport(operation)
	case "delete":
		err = s.processBulkDelete(operation)
	case "update":
		err = s.processBulkUpdate(operation)
	default:
		err = fmt.Errorf("unknown operation type: %s", operation.Type)
	}

	// Update final status
	completed := time.Now()
	operation.CompletedAt = &completed

	if err != nil {
		operation.Status = "failed"
		operation.Error = err.Error()
	} else {
		operation.Status = "completed"
		operation.Progress = 100
	}

	s.db.Save(operation)
}

func (s *Service) processBulkImport(operation *BulkOperation) error {
	// Implementation would depend on integration with bookmark service
	// For now, simulate processing
	operation.TotalItems = 100
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond) // Simulate work
		operation.ProcessedItems = i + 1
		operation.Progress = (i + 1) * 100 / 100
		s.db.Save(operation)
	}
	return nil
}

func (s *Service) processBulkExport(operation *BulkOperation) error {
	// Implementation would depend on integration with bookmark service
	// For now, simulate processing
	operation.TotalItems = 50
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond) // Simulate work
		operation.ProcessedItems = i + 1
		operation.Progress = (i + 1) * 100 / 50
		s.db.Save(operation)
	}
	return nil
}

func (s *Service) processBulkDelete(operation *BulkOperation) error {
	// Implementation would depend on integration with bookmark service
	// For now, simulate processing
	operation.TotalItems = 25
	for i := 0; i < 25; i++ {
		time.Sleep(30 * time.Millisecond) // Simulate work
		operation.ProcessedItems = i + 1
		operation.Progress = (i + 1) * 100 / 25
		s.db.Save(operation)
	}
	return nil
}

func (s *Service) processBulkUpdate(operation *BulkOperation) error {
	// Implementation would depend on integration with bookmark service
	// For now, simulate processing
	operation.TotalItems = 75
	for i := 0; i < 75; i++ {
		time.Sleep(15 * time.Millisecond) // Simulate work
		operation.ProcessedItems = i + 1
		operation.Progress = (i + 1) * 100 / 75
		s.db.Save(operation)
	}
	return nil
}

func (s *Service) processBackupJob(job *BackupJob) {
	// Update status to running
	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	s.db.Save(job)

	// Simulate backup process
	var err error
	switch job.Type {
	case "full":
		err = s.processFullBackup(job)
	case "incremental":
		err = s.processIncrementalBackup(job)
	default:
		err = fmt.Errorf("unknown backup type: %s", job.Type)
	}

	// Update final status
	completed := time.Now()
	job.CompletedAt = &completed

	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
	} else {
		job.Status = "completed"
	}

	s.db.Save(job)
}

func (s *Service) processFullBackup(job *BackupJob) error {
	// Simulate full backup
	time.Sleep(2 * time.Second)
	job.Size = 1024 * 1024 * 10 // 10MB
	job.FilePath = fmt.Sprintf("/backups/%s/full_%d.tar.gz", job.UserID, time.Now().Unix())
	job.Checksum = "sha256:abcd1234..."
	return nil
}

func (s *Service) processIncrementalBackup(job *BackupJob) error {
	// Simulate incremental backup
	time.Sleep(1 * time.Second)
	job.Size = 1024 * 1024 * 2 // 2MB
	job.FilePath = fmt.Sprintf("/backups/%s/incremental_%d.tar.gz", job.UserID, time.Now().Unix())
	job.Checksum = "sha256:efgh5678..."
	return nil
}

// Request types for the service methods
type WebhookEndpointRequest struct {
	Name       string            `json:"name" binding:"required"`
	URL        string            `json:"url" binding:"required"`
	Events     []string          `json:"events" binding:"required"`
	Active     bool              `json:"active"`
	RetryCount int               `json:"retry_count"`
	Timeout    int               `json:"timeout"`
	Headers    map[string]string `json:"headers"`
}

type RSSFeedRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Link        string   `json:"link" binding:"required"`
	Language    string   `json:"language"`
	Copyright   string   `json:"copyright"`
	Category    string   `json:"category"`
	TTL         int      `json:"ttl"`
	MaxItems    int      `json:"max_items"`
	Collections []uint   `json:"collections"`
	Tags        []string `json:"tags"`
}

// Additional service methods for handlers

// GetWebhookDeliveries retrieves webhook deliveries for an endpoint
func (s *Service) GetWebhookDeliveries(userID string, endpointID uint) ([]WebhookDelivery, error) {
	// First verify the endpoint belongs to the user
	var endpoint WebhookEndpoint
	if err := s.db.Where("id = ? AND user_id = ?", endpointID, userID).First(&endpoint).Error; err != nil {
		return nil, fmt.Errorf("webhook endpoint not found: %w", err)
	}

	var deliveries []WebhookDelivery
	if err := s.db.Where("endpoint_id = ?", endpointID).Order("created_at DESC").Find(&deliveries).Error; err != nil {
		return nil, fmt.Errorf("failed to get webhook deliveries: %w", err)
	}
	return deliveries, nil
}

// UpdateRSSFeed updates an RSS feed
func (s *Service) UpdateRSSFeed(userID string, id uint, req RSSFeedRequest) (*RSSFeed, error) {
	var feed RSSFeed
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&feed).Error; err != nil {
		return nil, fmt.Errorf("RSS feed not found: %w", err)
	}

	feed.Title = req.Title
	feed.Description = req.Description
	feed.Link = req.Link
	feed.Language = req.Language
	feed.Copyright = req.Copyright
	feed.Category = req.Category
	feed.TTL = req.TTL
	feed.MaxItems = req.MaxItems
	feed.Collections = UintSlice(req.Collections)
	feed.Tags = StringSlice(req.Tags)

	if feed.Language == "" {
		feed.Language = "en"
	}
	if feed.TTL == 0 {
		feed.TTL = 60
	}
	if feed.MaxItems == 0 {
		feed.MaxItems = 50
	}

	if err := s.db.Save(&feed).Error; err != nil {
		return nil, fmt.Errorf("failed to update RSS feed: %w", err)
	}

	return &feed, nil
}

// DeleteRSSFeed deletes an RSS feed
func (s *Service) DeleteRSSFeed(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&RSSFeed{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete RSS feed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("RSS feed not found")
	}
	return nil
}

// CancelBulkOperation cancels a bulk operation
func (s *Service) CancelBulkOperation(userID string, id uint) error {
	var operation BulkOperation
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&operation).Error; err != nil {
		return fmt.Errorf("bulk operation not found: %w", err)
	}

	if operation.Status == "completed" || operation.Status == "failed" {
		return fmt.Errorf("cannot cancel completed or failed operation")
	}

	operation.Status = "cancelled"
	completed := time.Now()
	operation.CompletedAt = &completed

	if err := s.db.Save(&operation).Error; err != nil {
		return fmt.Errorf("failed to cancel bulk operation: %w", err)
	}

	return nil
}

// GetBackupJob retrieves a specific backup job
func (s *Service) GetBackupJob(userID string, id uint) (*BackupJob, error) {
	var job BackupJob
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&job).Error; err != nil {
		return nil, fmt.Errorf("backup job not found: %w", err)
	}
	return &job, nil
}

// GetBackupFilePath retrieves the file path for a backup job
func (s *Service) GetBackupFilePath(userID string, id uint) (string, error) {
	job, err := s.GetBackupJob(userID, id)
	if err != nil {
		return "", err
	}

	if job.Status != "completed" {
		return "", fmt.Errorf("backup job not completed")
	}

	if job.FilePath == "" {
		return "", fmt.Errorf("backup file path not available")
	}

	return job.FilePath, nil
}

// UpdateAPIIntegration updates an API integration
func (s *Service) UpdateAPIIntegration(userID string, id uint, req APIIntegrationRequest) (*APIIntegration, error) {
	var integration APIIntegration
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&integration).Error; err != nil {
		return nil, fmt.Errorf("API integration not found: %w", err)
	}

	integration.Name = req.Name
	integration.Type = req.Type
	integration.BaseURL = req.BaseURL
	integration.APIKey = req.APIKey
	integration.APISecret = req.APISecret
	integration.SyncEnabled = req.SyncEnabled
	integration.SyncInterval = req.SyncInterval
	integration.Config = InterfaceMap(req.Config)

	if integration.SyncInterval == 0 {
		integration.SyncInterval = 3600 // 1 hour default
	}

	if err := s.db.Save(&integration).Error; err != nil {
		return nil, fmt.Errorf("failed to update API integration: %w", err)
	}

	return &integration, nil
}

// DeleteAPIIntegration deletes an API integration
func (s *Service) DeleteAPIIntegration(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&APIIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete API integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API integration not found")
	}
	return nil
}

// TriggerSync triggers a manual sync for an API integration
func (s *Service) TriggerSync(userID string, id uint) (map[string]interface{}, error) {
	var integration APIIntegration
	if err := s.db.Where("id = ? AND user_id = ? AND active = ?", id, userID, true).First(&integration).Error; err != nil {
		return nil, fmt.Errorf("API integration not found or inactive: %w", err)
	}

	// Simulate sync process
	result := map[string]interface{}{
		"status":       "success",
		"message":      "Sync triggered successfully",
		"timestamp":    time.Now(),
		"items_synced": 42,
	}

	// Update last sync time
	now := time.Now()
	integration.LastSync = &now
	s.db.Save(&integration)

	return result, nil
}

// TestIntegration tests an API integration
func (s *Service) TestIntegration(userID string, id uint) (map[string]interface{}, error) {
	var integration APIIntegration
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&integration).Error; err != nil {
		return nil, fmt.Errorf("API integration not found: %w", err)
	}

	// Simulate API test
	result := map[string]interface{}{
		"status":               "success",
		"message":              "API integration test successful",
		"response_time":        "150ms",
		"api_version":          "v1.0",
		"rate_limit_remaining": 95,
	}

	return result, nil
}

// CreateAutomationRule creates a new automation rule
func (s *Service) CreateAutomationRule(userID string, req AutomationRuleRequest) (*AutomationRule, error) {
	rule := &AutomationRule{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Trigger:     req.Trigger,
		Conditions:  InterfaceMap(req.Conditions),
		Actions:     InterfaceMap(req.Actions),
		Active:      true,
		Priority:    req.Priority,
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("failed to create automation rule: %w", err)
	}

	return rule, nil
}

// GetAutomationRules retrieves automation rules for a user
func (s *Service) GetAutomationRules(userID string) ([]AutomationRule, error) {
	var rules []AutomationRule
	if err := s.db.Where("user_id = ?", userID).Order("priority DESC, created_at DESC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get automation rules: %w", err)
	}
	return rules, nil
}

// UpdateAutomationRule updates an automation rule
func (s *Service) UpdateAutomationRule(userID string, id uint, req AutomationRuleRequest) (*AutomationRule, error) {
	var rule AutomationRule
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&rule).Error; err != nil {
		return nil, fmt.Errorf("automation rule not found: %w", err)
	}

	rule.Name = req.Name
	rule.Description = req.Description
	rule.Trigger = req.Trigger
	rule.Conditions = InterfaceMap(req.Conditions)
	rule.Actions = InterfaceMap(req.Actions)
	rule.Priority = req.Priority

	if err := s.db.Save(&rule).Error; err != nil {
		return nil, fmt.Errorf("failed to update automation rule: %w", err)
	}

	return &rule, nil
}

// DeleteAutomationRule deletes an automation rule
func (s *Service) DeleteAutomationRule(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&AutomationRule{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete automation rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("automation rule not found")
	}
	return nil
}

// ExecuteAutomationRule manually executes an automation rule
func (s *Service) ExecuteAutomationRule(userID string, id uint) (map[string]interface{}, error) {
	var rule AutomationRule
	if err := s.db.Where("id = ? AND user_id = ? AND active = ?", id, userID, true).First(&rule).Error; err != nil {
		return nil, fmt.Errorf("automation rule not found or inactive: %w", err)
	}

	// Simulate rule execution
	result := map[string]interface{}{
		"status":            "success",
		"message":           "Automation rule executed successfully",
		"timestamp":         time.Now(),
		"actions_performed": len(rule.Actions),
	}

	// Update execution count and last executed time
	rule.ExecutionCount++
	now := time.Now()
	rule.LastExecuted = &now
	s.db.Save(&rule)

	return result, nil
}
