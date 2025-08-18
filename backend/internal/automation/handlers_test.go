package automation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AutomationHandlerTestSuite defines the test suite for automation handlers
type AutomationHandlerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *Service
	handler *Handler
	router  *gin.Engine
	userID  string
}

// SetupSuite sets up the test suite
func (suite *AutomationHandlerTestSuite) SetupSuite() {
	suite.userID = "test-user-123"

	// Setup Gin router
	gin.SetMode(gin.TestMode)
}

// TearDownTest cleans up after each test
func (suite *AutomationHandlerTestSuite) TearDownTest() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

// SetupTest sets up each test with a fresh database
func (suite *AutomationHandlerTestSuite) SetupTest() {
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
	suite.handler = NewHandler(suite.service)

	// Setup Gin router for each test
	suite.router = gin.New()

	// Add middleware to set user_id
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", suite.userID)
		c.Next()
	})

	// Register routes
	api := suite.router.Group("/api/v1")
	suite.handler.RegisterRoutes(api)
}

// Helper method to make HTTP requests
func (suite *AutomationHandlerTestSuite) makeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// Webhook Endpoint Handler Tests

func (suite *AutomationHandlerTestSuite) TestCreateWebhookEndpoint_Success() {
	// Given: A valid webhook endpoint request
	reqBody := WebhookEndpointRequest{
		Name:       "Test Webhook",
		URL:        "https://example.com/webhook",
		Events:     []string{"bookmark.created", "bookmark.updated"},
		Active:     true,
		RetryCount: 3,
		Timeout:    30,
		Headers:    map[string]string{"Authorization": "Bearer token"},
	}

	// When: Making a POST request to create webhook endpoint
	w := suite.makeRequest("POST", "/api/v1/automation/webhooks", reqBody)

	// Then: The response should be successful
	suite.Equal(http.StatusCreated, w.Code)

	var response WebhookEndpoint
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(reqBody.Name, response.Name)
	suite.Equal(reqBody.URL, response.URL)
	suite.Equal(reqBody.Events, []string(response.Events))
	// Note: Secret field is hidden from JSON responses for security
}

func (suite *AutomationHandlerTestSuite) TestCreateWebhookEndpoint_InvalidRequest() {
	// Given: An invalid webhook endpoint request (missing required fields)
	reqBody := map[string]interface{}{
		"name": "Test Webhook",
		// Missing URL and Events
	}

	// When: Making a POST request with invalid data
	w := suite.makeRequest("POST", "/api/v1/automation/webhooks", reqBody)

	// Then: The response should be a bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "error")
}

func (suite *AutomationHandlerTestSuite) TestGetWebhookEndpoints_Success() {
	// Given: An existing webhook endpoint
	reqBody := WebhookEndpointRequest{
		Name:   "Test Webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"bookmark.created"},
	}
	suite.service.CreateWebhookEndpoint(suite.userID, reqBody)

	// When: Making a GET request to retrieve webhook endpoints
	w := suite.makeRequest("GET", "/api/v1/automation/webhooks", nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "endpoints")

	endpoints := response["endpoints"].([]interface{})
	suite.Len(endpoints, 1)
}

func (suite *AutomationHandlerTestSuite) TestUpdateWebhookEndpoint_Success() {
	// Given: An existing webhook endpoint
	createReq := WebhookEndpointRequest{
		Name:   "Original Webhook",
		URL:    "https://example.com/original",
		Events: []string{"bookmark.created"},
	}
	endpoint, _ := suite.service.CreateWebhookEndpoint(suite.userID, createReq)

	updateReq := WebhookEndpointRequest{
		Name:   "Updated Webhook",
		URL:    "https://example.com/updated",
		Events: []string{"bookmark.created", "bookmark.updated"},
		Active: false,
	}

	// When: Making a PUT request to update the webhook endpoint
	url := "/api/v1/automation/webhooks/" + strconv.Itoa(int(endpoint.ID))
	w := suite.makeRequest("PUT", url, updateReq)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response WebhookEndpoint
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(updateReq.Name, response.Name)
	suite.Equal(updateReq.URL, response.URL)
	suite.False(response.Active)
}

func (suite *AutomationHandlerTestSuite) TestDeleteWebhookEndpoint_Success() {
	// Given: An existing webhook endpoint
	reqBody := WebhookEndpointRequest{
		Name:   "Test Webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"bookmark.created"},
	}
	endpoint, _ := suite.service.CreateWebhookEndpoint(suite.userID, reqBody)

	// When: Making a DELETE request to delete the webhook endpoint
	url := "/api/v1/automation/webhooks/" + strconv.Itoa(int(endpoint.ID))
	w := suite.makeRequest("DELETE", url, nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "message")
}

// RSS Feed Handler Tests

func (suite *AutomationHandlerTestSuite) TestCreateRSSFeed_Success() {
	// Given: A valid RSS feed request
	reqBody := RSSFeedRequest{
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

	// When: Making a POST request to create RSS feed
	w := suite.makeRequest("POST", "/api/v1/automation/rss", reqBody)

	// Then: The response should be successful
	suite.Equal(http.StatusCreated, w.Code)

	var response RSSFeed
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(reqBody.Title, response.Title)
	suite.Equal(reqBody.Description, response.Description)
	suite.NotEmpty(response.PublicKey)
}

func (suite *AutomationHandlerTestSuite) TestGetPublicRSSFeed_Success() {
	// Given: An existing RSS feed
	reqBody := RSSFeedRequest{
		Title: "Test Feed",
		Link:  "https://example.com",
	}
	feed, _ := suite.service.CreateRSSFeed(suite.userID, reqBody)

	// When: Making a GET request to retrieve the public RSS feed
	url := "/api/v1/rss/" + feed.PublicKey
	w := suite.makeRequest("GET", url, nil)

	// Then: The response should be successful with RSS XML content
	suite.Equal(http.StatusOK, w.Code)
	suite.Equal("application/rss+xml; charset=utf-8", w.Header().Get("Content-Type"))
	suite.Contains(w.Body.String(), "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	suite.Contains(w.Body.String(), "<title><![CDATA[Test Feed]]></title>")
}

func (suite *AutomationHandlerTestSuite) TestGetPublicRSSFeed_NotFound() {
	// When: Making a GET request with a non-existent public key
	w := suite.makeRequest("GET", "/api/v1/rss/non-existent-key", nil)

	// Then: The response should be not found
	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "error")
}

// Bulk Operation Handler Tests

func (suite *AutomationHandlerTestSuite) TestCreateBulkOperation_Success() {
	// Given: A valid bulk operation request
	reqBody := BulkOperationRequest{
		Type: "import",
		Parameters: map[string]interface{}{
			"source": "chrome",
			"file":   "bookmarks.html",
		},
	}

	// When: Making a POST request to create bulk operation
	w := suite.makeRequest("POST", "/api/v1/automation/bulk", reqBody)

	// Then: The response should be successful
	suite.Equal(http.StatusCreated, w.Code)

	var response BulkOperation
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(reqBody.Type, response.Type)
	suite.Equal("pending", response.Status)
}

func (suite *AutomationHandlerTestSuite) TestGetBulkOperations_Success() {
	// Given: An existing bulk operation
	reqBody := BulkOperationRequest{
		Type:       "export",
		Parameters: map[string]interface{}{"format": "json"},
	}
	suite.service.CreateBulkOperation(suite.userID, reqBody)

	// When: Making a GET request to retrieve bulk operations
	w := suite.makeRequest("GET", "/api/v1/automation/bulk", nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "operations")

	operations := response["operations"].([]interface{})
	suite.Len(operations, 1)
}

func (suite *AutomationHandlerTestSuite) TestGetBulkOperation_Success() {
	// Given: An existing bulk operation
	reqBody := BulkOperationRequest{
		Type:       "import",
		Parameters: map[string]interface{}{"source": "firefox"},
	}
	operation, _ := suite.service.CreateBulkOperation(suite.userID, reqBody)

	// When: Making a GET request to retrieve the specific bulk operation
	url := "/api/v1/automation/bulk/" + strconv.Itoa(int(operation.ID))
	w := suite.makeRequest("GET", url, nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response BulkOperation
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(operation.ID, response.ID)
	suite.Equal(reqBody.Type, response.Type)
}

// Backup Job Handler Tests

func (suite *AutomationHandlerTestSuite) TestCreateBackupJob_Success() {
	// Given: A valid backup request
	reqBody := BackupRequest{
		Type:        "full",
		Compression: "gzip",
		Encrypted:   true,
	}

	// When: Making a POST request to create backup job
	w := suite.makeRequest("POST", "/api/v1/automation/backup", reqBody)

	// Then: The response should be successful
	suite.Equal(http.StatusCreated, w.Code)

	var response BackupJob
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(reqBody.Type, response.Type)
	suite.Equal("pending", response.Status)
	suite.True(response.Encrypted)
}

func (suite *AutomationHandlerTestSuite) TestGetBackupJobs_Success() {
	// Given: An existing backup job
	reqBody := BackupRequest{Type: "incremental"}
	suite.service.CreateBackupJob(suite.userID, reqBody)

	// When: Making a GET request to retrieve backup jobs
	w := suite.makeRequest("GET", "/api/v1/automation/backup", nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "jobs")

	jobs := response["jobs"].([]interface{})
	suite.Len(jobs, 1)
}

// API Integration Handler Tests

func (suite *AutomationHandlerTestSuite) TestCreateAPIIntegration_Success() {
	// Given: A valid API integration request
	reqBody := APIIntegrationRequest{
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

	// When: Making a POST request to create API integration
	w := suite.makeRequest("POST", "/api/v1/automation/integrations", reqBody)

	// Then: The response should be successful
	suite.Equal(http.StatusCreated, w.Code)

	var response APIIntegration
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(reqBody.Name, response.Name)
	suite.Equal(reqBody.Type, response.Type)
	suite.True(response.SyncEnabled)
}

func (suite *AutomationHandlerTestSuite) TestTriggerSync_Success() {
	// Given: An existing API integration
	reqBody := APIIntegrationRequest{
		Name:    "Test Integration",
		Type:    "pocket",
		BaseURL: "https://api.example.com",
		APIKey:  "test-key",
	}
	integration, _ := suite.service.CreateAPIIntegration(suite.userID, reqBody)

	// When: Making a POST request to trigger sync
	url := "/api/v1/automation/integrations/" + strconv.Itoa(int(integration.ID)) + "/sync"
	w := suite.makeRequest("POST", url, nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "result")

	result := response["result"].(map[string]interface{})
	suite.Equal("success", result["status"])
}

func (suite *AutomationHandlerTestSuite) TestTestIntegration_Success() {
	// Given: An existing API integration
	reqBody := APIIntegrationRequest{
		Name:    "Test Integration",
		Type:    "pocket",
		BaseURL: "https://api.example.com",
		APIKey:  "test-key",
	}
	integration, _ := suite.service.CreateAPIIntegration(suite.userID, reqBody)

	// When: Making a POST request to test integration
	url := "/api/v1/automation/integrations/" + strconv.Itoa(int(integration.ID)) + "/test"
	w := suite.makeRequest("POST", url, nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "result")

	result := response["result"].(map[string]interface{})
	suite.Equal("success", result["status"])
}

// Automation Rule Handler Tests

func (suite *AutomationHandlerTestSuite) TestCreateAutomationRule_Success() {
	// Given: A valid automation rule request
	reqBody := AutomationRuleRequest{
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

	// When: Making a POST request to create automation rule
	w := suite.makeRequest("POST", "/api/v1/automation/rules", reqBody)

	// Then: The response should be successful
	suite.Equal(http.StatusCreated, w.Code)

	var response AutomationRule
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(reqBody.Name, response.Name)
	suite.Equal(reqBody.Trigger, response.Trigger)
	suite.Equal(reqBody.Priority, response.Priority)
}

func (suite *AutomationHandlerTestSuite) TestExecuteAutomationRule_Success() {
	// Given: An existing automation rule
	reqBody := AutomationRuleRequest{
		Name:    "Test Rule",
		Trigger: "bookmark_added",
		Actions: map[string]interface{}{
			"add_tags": []string{"auto"},
		},
	}
	rule, _ := suite.service.CreateAutomationRule(suite.userID, reqBody)

	// When: Making a POST request to execute the rule
	url := "/api/v1/automation/rules/" + strconv.Itoa(int(rule.ID)) + "/execute"
	w := suite.makeRequest("POST", url, nil)

	// Then: The response should be successful
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "result")

	result := response["result"].(map[string]interface{})
	suite.Equal("success", result["status"])
}

// Error handling tests

func (suite *AutomationHandlerTestSuite) TestCreateWebhookEndpoint_Unauthorized() {
	// Given: A router without user authentication middleware
	router := gin.New()
	gin.SetMode(gin.TestMode)
	api := router.Group("/api/v1")
	suite.handler.RegisterRoutes(api)

	reqBody := WebhookEndpointRequest{
		Name:   "Test Webhook",
		URL:    "https://example.com/webhook",
		Events: []string{"bookmark.created"},
	}

	// When: Making a request without authentication
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/automation/webhooks", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Then: The response should be unauthorized
	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "error")
	suite.Equal("User not authenticated", response["error"])
}

func (suite *AutomationHandlerTestSuite) TestUpdateWebhookEndpoint_InvalidID() {
	// When: Making a PUT request with invalid ID
	reqBody := WebhookEndpointRequest{
		Name:   "Test",
		URL:    "https://example.com",
		Events: []string{"test"},
	}
	w := suite.makeRequest("PUT", "/api/v1/automation/webhooks/invalid-id", reqBody)

	// Then: The response should be bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "error")
	suite.Equal("Invalid endpoint ID", response["error"])
}

func (suite *AutomationHandlerTestSuite) TestGetBulkOperation_InvalidID() {
	// When: Making a GET request with invalid ID
	w := suite.makeRequest("GET", "/api/v1/automation/bulk/invalid-id", nil)

	// Then: The response should be bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "error")
	suite.Equal("Invalid operation ID", response["error"])
}

func (suite *AutomationHandlerTestSuite) TestGetPublicRSSFeed_EmptyKey() {
	// When: Making a GET request with empty public key
	w := suite.makeRequest("GET", "/api/v1/rss/", nil)

	// Then: The response should be not found (404) due to route not matching
	suite.Equal(http.StatusNotFound, w.Code)
}

// Run the test suite
func TestAutomationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AutomationHandlerTestSuite))
}

// Additional individual tests for edge cases

func TestNewHandler(t *testing.T) {
	// Given: A service
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	service := NewService(db)

	// When: Creating a new handler
	handler := NewHandler(service)

	// Then: Handler should be initialized correctly
	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.service)
}

func TestRegisterRoutes(t *testing.T) {
	// Given: A handler and router
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	service := NewService(db)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")

	// When: Registering routes
	handler.RegisterRoutes(api)

	// Then: Routes should be registered (we can test this by checking if routes exist)
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Check for some key routes
	routePaths := make([]string, len(routes))
	for i, route := range routes {
		routePaths[i] = route.Path
	}

	assert.Contains(t, routePaths, "/api/v1/automation/webhooks")
	assert.Contains(t, routePaths, "/api/v1/automation/rss")
	assert.Contains(t, routePaths, "/api/v1/automation/bulk")
	assert.Contains(t, routePaths, "/api/v1/automation/backup")
	assert.Contains(t, routePaths, "/api/v1/automation/integrations")
	assert.Contains(t, routePaths, "/api/v1/automation/rules")
	assert.Contains(t, routePaths, "/api/v1/rss/:publicKey")
}
