package automation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AutomationServiceTestSuite defines the test suite for automation service
type AutomationServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *Service
	userID  string
}

// SetupSuite sets up the test suite
func (suite *AutomationServiceTestSuite) SetupSuite() {
	suite.userID = "test-user-123"
}

// TearDownTest cleans up after each test
func (suite *AutomationServiceTestSuite) TearDownTest() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

// SetupTest sets up each test with a fresh database
func (suite *AutomationServiceTestSuite) SetupTest() {
	// Create a new in-memory SQLite database for each test
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&WebhookEndpoint{},
		&WebhookDelivery{},
		&RSSFeed{},
		&BulkOperation{},
		&BackupJob{},
		&APIIntegration{},
		&AutomationRule{},
	)
	suite.Require().NoError(err)

	suite.db = db
	suite.service = NewServiceForTesting(db)
}

// Webhook Endpoint Tests

func (suite *AutomationServiceTestSuite) TestCreateWebhookEndpoint_Success() {
	// Given: A valid webhook endpoint request
	req := WebhookEndpointRequest{
		Name:       "Test Webhook",
		URL:        "https://example.com/webhook",
		Events:     []string{"bookmark.created", "bookmark.updated"},
		Active:     true,
		RetryCount: 3,
		Timeout:    30,
		Headers:    map[string]string{"Authorization": "Bearer token"},
	}

	// When: Creating a webhook endpoint
	endpoint, err := suite.service.CreateWebhookEndpoint(suite.userID, req)

	// Then: The endpoint should be created successfully
	suite.NoError(err)
	suite.NotNil(endpoint)
	suite.Equal(req.Name, endpoint.Name)
	suite.Equal(req.URL, endpoint.URL)
	suite.Equal(StringSlice(req.Events), endpoint.Events)
	suite.True(endpoint.Active)
	suite.Equal(3, endpoint.RetryCount)
	suite.Equal(30, endpoint.Timeout)
	suite.NotEmpty(endpoint.Secret)
	suite.Equal(suite.userID, endpoint.UserID)
}

func (suite *AutomationServiceTestSuite) TestCreateWebhookEndpoint_DefaultValues() {
	// Given: A webhook endpoint request with minimal data
	req := WebhookEndpointRequest{
		Name:   "Minimal Webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"bookmark.created"},
	}

	// When: Creating a webhook endpoint
	endpoint, err := suite.service.CreateWebhookEndpoint(suite.userID, req)

	// Then: Default values should be applied
	suite.NoError(err)
	suite.Equal(3, endpoint.RetryCount) // Default retry count
	suite.Equal(30, endpoint.Timeout)   // Default timeout
	suite.True(endpoint.Active)         // Default active state
}

func (suite *AutomationServiceTestSuite) TestGetWebhookEndpoints_Success() {
	// Given: Multiple webhook endpoints for the user
	req1 := WebhookEndpointRequest{
		Name:   "Webhook 1",
		URL:    "https://example.com/webhook1",
		Events: []string{"bookmark.created"},
	}
	req2 := WebhookEndpointRequest{
		Name:   "Webhook 2",
		URL:    "https://example.com/webhook2",
		Events: []string{"bookmark.updated"},
	}

	endpoint1, _ := suite.service.CreateWebhookEndpoint(suite.userID, req1)
	endpoint2, _ := suite.service.CreateWebhookEndpoint(suite.userID, req2)

	// When: Getting webhook endpoints
	endpoints, err := suite.service.GetWebhookEndpoints(suite.userID)

	// Then: All endpoints should be returned
	suite.NoError(err)
	suite.Len(endpoints, 2)

	// Verify endpoints are returned
	endpointIDs := []uint{endpoints[0].ID, endpoints[1].ID}
	suite.Contains(endpointIDs, endpoint1.ID)
	suite.Contains(endpointIDs, endpoint2.ID)
}

func (suite *AutomationServiceTestSuite) TestUpdateWebhookEndpoint_Success() {
	// Given: An existing webhook endpoint
	createReq := WebhookEndpointRequest{
		Name:   "Original Webhook",
		URL:    "https://example.com/original",
		Events: []string{"bookmark.created"},
	}
	endpoint, _ := suite.service.CreateWebhookEndpoint(suite.userID, createReq)

	// When: Updating the webhook endpoint
	updateReq := WebhookEndpointRequest{
		Name:   "Updated Webhook",
		URL:    "https://example.com/updated",
		Events: []string{"bookmark.created", "bookmark.updated"},
		Active: false,
	}
	updatedEndpoint, err := suite.service.UpdateWebhookEndpoint(suite.userID, endpoint.ID, updateReq)

	// Then: The endpoint should be updated
	suite.NoError(err)
	suite.Equal(updateReq.Name, updatedEndpoint.Name)
	suite.Equal(updateReq.URL, updatedEndpoint.URL)
	suite.Equal(StringSlice(updateReq.Events), updatedEndpoint.Events)
	suite.False(updatedEndpoint.Active)
}

func (suite *AutomationServiceTestSuite) TestDeleteWebhookEndpoint_Success() {
	// Given: An existing webhook endpoint
	req := WebhookEndpointRequest{
		Name:   "Test Webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"bookmark.created"},
	}
	endpoint, _ := suite.service.CreateWebhookEndpoint(suite.userID, req)

	// When: Deleting the webhook endpoint
	err := suite.service.DeleteWebhookEndpoint(suite.userID, endpoint.ID)

	// Then: The endpoint should be deleted
	suite.NoError(err)

	// Verify endpoint is deleted
	endpoints, _ := suite.service.GetWebhookEndpoints(suite.userID)
	suite.Len(endpoints, 0)
}

func (suite *AutomationServiceTestSuite) TestTriggerWebhook_Success() {
	// Given: An active webhook endpoint subscribed to bookmark.created event
	req := WebhookEndpointRequest{
		Name:   "Test Webhook",
		URL:    "https://httpbin.org/post", // Using httpbin for testing
		Events: []string{"bookmark.created"},
		Active: true,
	}
	endpoint, _ := suite.service.CreateWebhookEndpoint(suite.userID, req)

	// When: Triggering a webhook
	testData := map[string]interface{}{
		"bookmark_id": "123",
		"title":       "Test Bookmark",
		"url":         "https://example.com",
	}

	ctx := context.Background()
	err := suite.service.TriggerWebhook(ctx, WebhookEventBookmarkCreated, suite.userID, testData)

	// Then: The webhook should be triggered without error
	suite.NoError(err)

	// Verify delivery record is created
	time.Sleep(100 * time.Millisecond) // Allow time for async processing
	deliveries, _ := suite.service.GetWebhookDeliveries(suite.userID, endpoint.ID)
	suite.Len(deliveries, 1)
	suite.Equal("pending", deliveries[0].Status)
	suite.Equal(WebhookEventBookmarkCreated, deliveries[0].Event)
}

// RSS Feed Tests

func (suite *AutomationServiceTestSuite) TestCreateRSSFeed_Success() {
	// Given: A valid RSS feed request
	req := RSSFeedRequest{
		Title:       "My Bookmarks",
		Description: "My personal bookmark collection",
		Link:        "https://example.com",
		Language:    "en",
		Copyright:   "© 2024 Test User",
		Category:    "Technology",
		TTL:         60,
		MaxItems:    50,
		Collections: []uint{1, 2, 3},
		Tags:        []string{"tech", "programming"},
	}

	// When: Creating an RSS feed
	feed, err := suite.service.CreateRSSFeed(suite.userID, req)

	// Then: The feed should be created successfully
	suite.NoError(err)
	suite.NotNil(feed)
	suite.Equal(req.Title, feed.Title)
	suite.Equal(req.Description, feed.Description)
	suite.Equal(req.Link, feed.Link)
	suite.Equal(req.Language, feed.Language)
	suite.Equal(req.Copyright, feed.Copyright)
	suite.Equal(req.Category, feed.Category)
	suite.Equal(req.TTL, feed.TTL)
	suite.Equal(req.MaxItems, feed.MaxItems)
	suite.Equal(UintSlice(req.Collections), feed.Collections)
	suite.Equal(StringSlice(req.Tags), feed.Tags)
	suite.NotEmpty(feed.PublicKey)
	suite.True(feed.Active)
}

func (suite *AutomationServiceTestSuite) TestCreateRSSFeed_DefaultValues() {
	// Given: An RSS feed request with minimal data
	req := RSSFeedRequest{
		Title: "Minimal Feed",
		Link:  "https://example.com",
	}

	// When: Creating an RSS feed
	feed, err := suite.service.CreateRSSFeed(suite.userID, req)

	// Then: Default values should be applied
	suite.NoError(err)
	suite.Equal("en", feed.Language) // Default language
	suite.Equal(60, feed.TTL)        // Default TTL
	suite.Equal(50, feed.MaxItems)   // Default max items
	suite.True(feed.Active)          // Default active state
}

func (suite *AutomationServiceTestSuite) TestGetRSSFeedByPublicKey_Success() {
	// Given: An existing RSS feed
	req := RSSFeedRequest{
		Title: "Test Feed",
		Link:  "https://example.com",
	}
	feed, _ := suite.service.CreateRSSFeed(suite.userID, req)

	// When: Getting the RSS feed by public key
	retrievedFeed, err := suite.service.GetRSSFeedByPublicKey(feed.PublicKey)

	// Then: The feed should be retrieved successfully
	suite.NoError(err)
	suite.Equal(feed.ID, retrievedFeed.ID)
	suite.Equal(feed.Title, retrievedFeed.Title)
	suite.Equal(feed.PublicKey, retrievedFeed.PublicKey)
}

func (suite *AutomationServiceTestSuite) TestGenerateRSSContent_Success() {
	// Given: An RSS feed
	req := RSSFeedRequest{
		Title:       "Test Feed",
		Description: "Test Description",
		Link:        "https://example.com",
		Language:    "en",
		Copyright:   "© 2024 Test",
		Category:    "Technology",
		TTL:         60,
	}
	feed, _ := suite.service.CreateRSSFeed(suite.userID, req)

	// When: Generating RSS content
	content, err := suite.service.GenerateRSSContent(feed)

	// Then: Valid RSS XML should be generated
	suite.NoError(err)
	suite.Contains(content, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	suite.Contains(content, "<rss version=\"2.0\"")
	suite.Contains(content, "<title><![CDATA[Test Feed]]></title>")
	suite.Contains(content, "<description><![CDATA[Test Description]]></description>")
	suite.Contains(content, "<link>https://example.com</link>")
	suite.Contains(content, "<language>en</language>")
	suite.Contains(content, "<ttl>60</ttl>")
}

// Bulk Operation Tests

func (suite *AutomationServiceTestSuite) TestCreateBulkOperation_Success() {
	// Given: A valid bulk operation request
	req := BulkOperationRequest{
		Type: "import",
		Parameters: map[string]interface{}{
			"source": "chrome",
			"file":   "bookmarks.html",
		},
	}

	// When: Creating a bulk operation
	operation, err := suite.service.CreateBulkOperation(suite.userID, req)

	// Then: The operation should be created successfully
	suite.NoError(err)
	suite.NotNil(operation)
	suite.Equal(req.Type, operation.Type)
	// Status could be "pending" or "running" due to async processing
	suite.Contains([]string{"pending", "running"}, operation.Status)
	suite.Equal(InterfaceMap(req.Parameters), operation.Parameters)
	suite.Equal(suite.userID, operation.UserID)
}

func (suite *AutomationServiceTestSuite) TestGetBulkOperations_Success() {
	// Given: Multiple bulk operations for the user
	req1 := BulkOperationRequest{Type: "import", Parameters: map[string]interface{}{"source": "chrome"}}
	req2 := BulkOperationRequest{Type: "export", Parameters: map[string]interface{}{"format": "json"}}

	operation1, _ := suite.service.CreateBulkOperation(suite.userID, req1)
	operation2, _ := suite.service.CreateBulkOperation(suite.userID, req2)

	// When: Getting bulk operations
	operations, err := suite.service.GetBulkOperations(suite.userID)

	// Then: All operations should be returned
	suite.NoError(err)
	suite.Len(operations, 2)

	// Verify operations are returned (ordered by created_at DESC)
	suite.Equal(operation2.ID, operations[0].ID) // Most recent first
	suite.Equal(operation1.ID, operations[1].ID)
}

func (suite *AutomationServiceTestSuite) TestCancelBulkOperation_Success() {
	// Given: A pending bulk operation
	req := BulkOperationRequest{Type: "import", Parameters: map[string]interface{}{}}
	operation, _ := suite.service.CreateBulkOperation(suite.userID, req)

	// When: Cancelling the bulk operation
	err := suite.service.CancelBulkOperation(suite.userID, operation.ID)

	// Then: The operation should be cancelled
	suite.NoError(err)

	// Verify operation status is updated
	updatedOperation, _ := suite.service.GetBulkOperation(suite.userID, operation.ID)
	suite.Equal("cancelled", updatedOperation.Status)
	suite.NotNil(updatedOperation.CompletedAt)
}

func (suite *AutomationServiceTestSuite) TestCancelBulkOperation_AlreadyCompleted() {
	// Given: A completed bulk operation
	req := BulkOperationRequest{Type: "import", Parameters: map[string]interface{}{}}
	operation, _ := suite.service.CreateBulkOperation(suite.userID, req)

	// Manually set status to completed
	operation.Status = "completed"
	suite.db.Save(operation)

	// When: Attempting to cancel the completed operation
	err := suite.service.CancelBulkOperation(suite.userID, operation.ID)

	// Then: An error should be returned
	suite.Error(err)
	suite.Contains(err.Error(), "cannot cancel completed or failed operation")
}

// Backup Job Tests

func (suite *AutomationServiceTestSuite) TestCreateBackupJob_Success() {
	// Given: A valid backup request
	req := BackupRequest{
		Type:        "full",
		Compression: "gzip",
		Encrypted:   true,
	}

	// When: Creating a backup job
	job, err := suite.service.CreateBackupJob(suite.userID, req)

	// Then: The job should be created successfully
	suite.NoError(err)
	suite.NotNil(job)
	suite.Equal(req.Type, job.Type)
	// Status could be "pending" or "running" due to async processing
	suite.Contains([]string{"pending", "running"}, job.Status)
	suite.Equal(req.Compression, job.Compression)
	suite.True(job.Encrypted)
	suite.Equal(30, job.RetentionDays) // Default retention
	suite.Equal(suite.userID, job.UserID)
}

func (suite *AutomationServiceTestSuite) TestCreateBackupJob_DefaultCompression() {
	// Given: A backup request without compression specified
	req := BackupRequest{
		Type: "incremental",
	}

	// When: Creating a backup job
	job, err := suite.service.CreateBackupJob(suite.userID, req)

	// Then: Default compression should be applied
	suite.NoError(err)
	suite.Equal("gzip", job.Compression) // Default compression
}

func (suite *AutomationServiceTestSuite) TestGetBackupJobs_Success() {
	// Given: Multiple backup jobs for the user
	req1 := BackupRequest{Type: "full"}
	req2 := BackupRequest{Type: "incremental"}

	job1, _ := suite.service.CreateBackupJob(suite.userID, req1)
	job2, _ := suite.service.CreateBackupJob(suite.userID, req2)

	// When: Getting backup jobs
	jobs, err := suite.service.GetBackupJobs(suite.userID)

	// Then: All jobs should be returned
	suite.NoError(err)
	suite.Len(jobs, 2)

	// Verify jobs are returned (ordered by created_at DESC)
	jobIDs := []uint{jobs[0].ID, jobs[1].ID}
	suite.Contains(jobIDs, job1.ID)
	suite.Contains(jobIDs, job2.ID)
}

// API Integration Tests

func (suite *AutomationServiceTestSuite) TestCreateAPIIntegration_Success() {
	// Given: A valid API integration request
	req := APIIntegrationRequest{
		Name:         "Pocket Integration",
		Type:         "pocket",
		BaseURL:      "https://getpocket.com/v3",
		APIKey:       "test-api-key",
		APISecret:    "test-api-secret",
		SyncEnabled:  true,
		SyncInterval: 3600,
		Config: map[string]interface{}{
			"consumer_key": "test-consumer-key",
		},
	}

	// When: Creating an API integration
	integration, err := suite.service.CreateAPIIntegration(suite.userID, req)

	// Then: The integration should be created successfully
	suite.NoError(err)
	suite.NotNil(integration)
	suite.Equal(req.Name, integration.Name)
	suite.Equal(req.Type, integration.Type)
	suite.Equal(req.BaseURL, integration.BaseURL)
	suite.Equal(req.APIKey, integration.APIKey)
	suite.Equal(req.APISecret, integration.APISecret)
	suite.True(integration.SyncEnabled)
	suite.Equal(req.SyncInterval, integration.SyncInterval)
	suite.Equal(InterfaceMap(req.Config), integration.Config)
	suite.True(integration.Active)
	suite.Equal(100, integration.RateLimit) // Default rate limit
}

func (suite *AutomationServiceTestSuite) TestTriggerSync_Success() {
	// Given: An active API integration
	req := APIIntegrationRequest{
		Name:    "Test Integration",
		Type:    "pocket",
		BaseURL: "https://api.example.com",
		APIKey:  "test-key",
	}
	integration, _ := suite.service.CreateAPIIntegration(suite.userID, req)

	// When: Triggering a sync
	result, err := suite.service.TriggerSync(suite.userID, integration.ID)

	// Then: The sync should be triggered successfully
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("success", result["status"])
	suite.Contains(result, "items_synced")

	// Verify last sync time is updated
	updatedIntegration, _ := suite.service.GetAPIIntegrations(suite.userID)
	suite.NotNil(updatedIntegration[0].LastSync)
}

func (suite *AutomationServiceTestSuite) TestTestIntegration_Success() {
	// Given: An API integration
	req := APIIntegrationRequest{
		Name:    "Test Integration",
		Type:    "pocket",
		BaseURL: "https://api.example.com",
		APIKey:  "test-key",
	}
	integration, _ := suite.service.CreateAPIIntegration(suite.userID, req)

	// When: Testing the integration
	result, err := suite.service.TestIntegration(suite.userID, integration.ID)

	// Then: The test should be successful
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("success", result["status"])
	suite.Contains(result, "response_time")
	suite.Contains(result, "api_version")
}

// Automation Rule Tests

func (suite *AutomationServiceTestSuite) TestCreateAutomationRule_Success() {
	// Given: A valid automation rule request
	req := AutomationRuleRequest{
		Name:        "Auto-tag Tech Bookmarks",
		Description: "Automatically tag bookmarks from tech websites",
		Trigger:     "bookmark_added",
		Conditions: map[string]interface{}{
			"url_contains": []string{"github.com", "stackoverflow.com"},
		},
		Actions: map[string]interface{}{
			"add_tags": []string{"tech", "programming"},
		},
		Priority: 10,
	}

	// When: Creating an automation rule
	rule, err := suite.service.CreateAutomationRule(suite.userID, req)

	// Then: The rule should be created successfully
	suite.NoError(err)
	suite.NotNil(rule)
	suite.Equal(req.Name, rule.Name)
	suite.Equal(req.Description, rule.Description)
	suite.Equal(req.Trigger, rule.Trigger)
	suite.Equal(InterfaceMap(req.Conditions), rule.Conditions)
	suite.Equal(InterfaceMap(req.Actions), rule.Actions)
	suite.Equal(req.Priority, rule.Priority)
	suite.True(rule.Active)
	suite.Equal(0, rule.ExecutionCount)
}

func (suite *AutomationServiceTestSuite) TestExecuteAutomationRule_Success() {
	// Given: An active automation rule
	req := AutomationRuleRequest{
		Name:    "Test Rule",
		Trigger: "bookmark_added",
		Actions: map[string]interface{}{
			"add_tags": []string{"auto"},
		},
	}
	rule, _ := suite.service.CreateAutomationRule(suite.userID, req)

	// When: Executing the automation rule
	result, err := suite.service.ExecuteAutomationRule(suite.userID, rule.ID)

	// Then: The rule should be executed successfully
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("success", result["status"])
	suite.Contains(result, "actions_performed")

	// Verify execution count is updated
	updatedRule, _ := suite.service.GetAutomationRules(suite.userID)
	suite.Equal(1, updatedRule[0].ExecutionCount)
	suite.NotNil(updatedRule[0].LastExecuted)
}

// Helper method tests

func (suite *AutomationServiceTestSuite) TestGenerateSecret_Success() {
	// When: Generating a secret
	secret, err := suite.service.generateSecret()

	// Then: A valid secret should be generated
	suite.NoError(err)
	suite.NotEmpty(secret)
	suite.Len(secret, 64) // 32 bytes = 64 hex characters
}

func (suite *AutomationServiceTestSuite) TestGeneratePublicKey_Success() {
	// When: Generating a public key
	publicKey, err := suite.service.generatePublicKey()

	// Then: A valid public key should be generated
	suite.NoError(err)
	suite.NotEmpty(publicKey)
	suite.Len(publicKey, 32) // 16 bytes = 32 hex characters
}

func (suite *AutomationServiceTestSuite) TestIsEventSubscribed_Success() {
	// Given: A list of subscribed events
	events := StringSlice{"bookmark.created", "bookmark.updated", "collection.created"}

	// When: Checking if events are subscribed
	isSubscribed1 := suite.service.isEventSubscribed(events, "bookmark.created")
	isSubscribed2 := suite.service.isEventSubscribed(events, "bookmark.deleted")

	// Then: Correct subscription status should be returned
	suite.True(isSubscribed1)
	suite.False(isSubscribed2)
}

// Run the test suite
func TestAutomationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AutomationServiceTestSuite))
}

// Additional individual tests for edge cases

func TestNewService(t *testing.T) {
	// Given: A database connection
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// When: Creating a new service
	service := NewService(db)

	// Then: Service should be initialized correctly
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.NotNil(t, service.httpClient)
	assert.Equal(t, 30*time.Second, service.httpClient.Timeout)
}

func TestWebhookEndpoint_NotFound(t *testing.T) {
	// Given: A service with empty database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&WebhookEndpoint{})
	service := NewService(db)

	// When: Trying to update a non-existent webhook endpoint
	req := WebhookEndpointRequest{
		Name:   "Test",
		URL:    "https://example.com",
		Events: []string{"test"},
	}
	_, err := service.UpdateWebhookEndpoint("user123", 999, req)

	// Then: An error should be returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook endpoint not found")
}

func TestRSSFeed_NotFound(t *testing.T) {
	// Given: A service with empty database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&RSSFeed{})
	service := NewService(db)

	// When: Trying to get a non-existent RSS feed by public key
	_, err := service.GetRSSFeedByPublicKey("non-existent-key")

	// Then: An error should be returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RSS feed not found")
}

func TestBulkOperation_NotFound(t *testing.T) {
	// Given: A service with empty database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&BulkOperation{})
	service := NewService(db)

	// When: Trying to get a non-existent bulk operation
	_, err := service.GetBulkOperation("user123", 999)

	// Then: An error should be returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bulk operation not found")
}

func TestBackupJob_NotCompleted(t *testing.T) {
	// Given: A service with a pending backup job
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&BackupJob{})
	service := NewService(db)

	job := &BackupJob{
		UserID: "user123",
		Type:   "full",
		Status: "pending",
	}
	db.Create(job)

	// When: Trying to get the file path for a non-completed job
	_, err := service.GetBackupFilePath("user123", job.ID)

	// Then: An error should be returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup job not completed")
}

func TestAPIIntegration_NotFound(t *testing.T) {
	// Given: A service with empty database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&APIIntegration{})
	service := NewService(db)

	// When: Trying to trigger sync for a non-existent integration
	_, err := service.TriggerSync("user123", 999)

	// Then: An error should be returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API integration not found")
}

func TestAutomationRule_NotFound(t *testing.T) {
	// Given: A service with empty database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&AutomationRule{})
	service := NewService(db)

	// When: Trying to execute a non-existent rule
	_, err := service.ExecuteAutomationRule("user123", 999)

	// Then: An error should be returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "automation rule not found")
}
