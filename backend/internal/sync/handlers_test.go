package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SyncHandlerTestSuite defines the test suite for sync handlers
type SyncHandlerTestSuite struct {
	suite.Suite
	handler     *Handler
	service     *Service
	db          *gorm.DB
	redisClient *MockRedisClient
	router      *gin.Engine
	logger      *zap.Logger
}

func (suite *SyncHandlerTestSuite) SetupTest() {
	// Setup test database
	db, err := database.SetupTestDB()
	suite.Require().NoError(err)
	suite.db = db

	// Setup mock Redis client
	suite.redisClient = &MockRedisClient{}

	// Setup logger
	suite.logger = zap.NewNop()

	// Create service and handler
	suite.service = NewService(suite.db, suite.redisClient, suite.logger)
	suite.handler = NewHandler(suite.service, suite.logger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware to set user_id
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	{
		sync := v1.Group("/sync")
		{
			sync.GET("/state", suite.handler.GetSyncState)
			sync.PUT("/state", suite.handler.UpdateSyncState)
			sync.GET("/delta", suite.handler.GetDeltaSync)
			sync.POST("/events", suite.handler.CreateSyncEvent)
			sync.GET("/offline-queue", suite.handler.GetOfflineQueue)
			sync.POST("/offline-queue", suite.handler.QueueOfflineEvent)
			sync.POST("/offline-queue/process", suite.handler.ProcessOfflineQueue)
		}
	}
}

func (suite *SyncHandlerTestSuite) TearDownTest() {
	database.CleanupTestDB(suite.db)
}

func TestSyncHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SyncHandlerTestSuite))
}

// Test GET /api/v1/sync/state
func (suite *SyncHandlerTestSuite) TestGetSyncState() {
	// Test missing device_id
	req, _ := http.NewRequest("GET", "/api/v1/sync/state", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	// Test successful request
	req, _ = http.NewRequest("GET", "/api/v1/sync/state?device_id=device-123", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])

	data := response["data"].(map[string]interface{})
	suite.Equal("test-user-123", data["user_id"])
	suite.Equal("device-123", data["device_id"])
}

// Test PUT /api/v1/sync/state
func (suite *SyncHandlerTestSuite) TestUpdateSyncState() {
	requestBody := map[string]interface{}{
		"device_id":      "device-123",
		"last_sync_time": time.Now().Unix(),
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/api/v1/sync/state", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])
}

// Test GET /api/v1/sync/delta
func (suite *SyncHandlerTestSuite) TestGetDeltaSync() {
	// Create some test events
	event := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-123",
		Action:     "create",
		DeviceID:   "other-device",
		Timestamp:  time.Now().Add(-30 * time.Minute),
	}
	err := suite.db.Create(event).Error
	suite.NoError(err)

	// Test missing device_id
	req, _ := http.NewRequest("GET", "/api/v1/sync/delta", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	// Test successful request
	lastSyncTime := time.Now().Add(-1 * time.Hour).Unix()
	url := fmt.Sprintf("/api/v1/sync/delta?device_id=device-123&last_sync_time=%d", lastSyncTime)
	req, _ = http.NewRequest("GET", url, nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])

	data := response["data"].(map[string]interface{})
	events := data["events"].([]interface{})
	suite.Len(events, 1)
}

// Test POST /api/v1/sync/events
func (suite *SyncHandlerTestSuite) TestCreateSyncEvent() {
	suite.redisClient.On("PublishSyncEvent", mock.Anything, "test-user-123", mock.AnythingOfType("*sync.SyncEvent")).Return(nil)

	requestBody := map[string]interface{}{
		"type":        "bookmark_created",
		"resource_id": "bookmark-123",
		"action":      "create",
		"device_id":   "device-123",
		"data": map[string]interface{}{
			"title": "Test Bookmark",
			"url":   "https://example.com",
		},
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/sync/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])

	// Verify event was created in database
	var event SyncEvent
	err = suite.db.Where("user_id = ? AND resource_id = ?", "test-user-123", "bookmark-123").First(&event).Error
	suite.NoError(err)
	suite.Equal(SyncEventBookmarkCreated, event.Type)

	suite.redisClient.AssertExpected(suite.T())
}

// Test GET /api/v1/sync/offline-queue
func (suite *SyncHandlerTestSuite) TestGetOfflineQueue() {
	// Create offline events
	event := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-offline",
		Action:     "create",
		DeviceID:   "device-123",
		Status:     SyncStatusPending,
		Timestamp:  time.Now(),
	}
	err := suite.db.Create(event).Error
	suite.NoError(err)

	// Test missing device_id
	req, _ := http.NewRequest("GET", "/api/v1/sync/offline-queue", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	// Test successful request
	req, _ = http.NewRequest("GET", "/api/v1/sync/offline-queue?device_id=device-123", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])

	data := response["data"].(map[string]interface{})
	events := data["events"].([]interface{})
	suite.Len(events, 1)
}

// Test POST /api/v1/sync/offline-queue
func (suite *SyncHandlerTestSuite) TestQueueOfflineEvent() {
	requestBody := map[string]interface{}{
		"type":        "bookmark_created",
		"resource_id": "bookmark-offline",
		"action":      "create",
		"device_id":   "device-123",
		"data": map[string]interface{}{
			"title": "Offline Bookmark",
			"url":   "https://offline.com",
		},
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/sync/offline-queue", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])

	// Verify event was queued in database
	var event SyncEvent
	err = suite.db.Where("user_id = ? AND resource_id = ? AND status = ?", "test-user-123", "bookmark-offline", SyncStatusPending).First(&event).Error
	suite.NoError(err)
	suite.Equal(SyncEventBookmarkCreated, event.Type)
}

// Test POST /api/v1/sync/offline-queue/process
func (suite *SyncHandlerTestSuite) TestProcessOfflineQueue() {
	// Create offline events
	event := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-offline",
		Action:     "create",
		DeviceID:   "device-123",
		Status:     SyncStatusPending,
		Timestamp:  time.Now(),
	}
	err := suite.db.Create(event).Error
	suite.NoError(err)

	suite.redisClient.On("PublishSyncEvent", mock.Anything, "test-user-123", mock.AnythingOfType("*sync.SyncEvent")).Return(nil)

	requestBody := map[string]interface{}{
		"device_id": "device-123",
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/sync/offline-queue/process", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(true, response["success"])

	// Verify event status was updated
	var processedEvent SyncEvent
	err = suite.db.Where("user_id = ? AND resource_id = ?", "test-user-123", "bookmark-offline").First(&processedEvent).Error
	suite.NoError(err)
	suite.Equal(SyncStatusSynced, processedEvent.Status)

	suite.redisClient.AssertExpected(suite.T())
}
