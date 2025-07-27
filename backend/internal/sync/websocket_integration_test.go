package sync

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/websocket"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebSocketSyncIntegrationTestSuite tests sync service integration for WebSocket features
type WebSocketSyncIntegrationTestSuite struct {
	suite.Suite
	service     *Service
	db          *gorm.DB
	redisClient *MockRedisClient
	logger      *zap.Logger
}

func (suite *WebSocketSyncIntegrationTestSuite) SetupTest() {
	// Setup test database
	db, err := database.SetupTestDB()
	suite.Require().NoError(err)
	suite.db = db

	// Setup mock Redis client
	suite.redisClient = &MockRedisClient{}

	// Setup logger
	suite.logger = zap.NewNop()

	// Create sync service
	suite.service = NewService(suite.db, suite.redisClient, suite.logger)
}

func (suite *WebSocketSyncIntegrationTestSuite) TearDownTest() {
	database.CleanupTestDB(suite.db)
}

func TestWebSocketSyncIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(WebSocketSyncIntegrationTestSuite))
}

// Test sync message handling for WebSocket integration
func (suite *WebSocketSyncIntegrationTestSuite) TestSyncMessageHandling() {
	// Test ping message
	pingMsg := &websocket.SyncMessage{
		Type:      "ping",
		UserID:    "test-user-123",
		DeviceID:  "device-456",
		Timestamp: time.Now(),
	}

	response, err := suite.service.HandleSyncMessage(context.Background(), pingMsg)
	suite.NoError(err)
	suite.Equal("pong", response.Type)
}

// Test sync request handling for WebSocket
func (suite *WebSocketSyncIntegrationTestSuite) TestSyncRequestHandling() {
	// Create some test sync events
	event := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-123",
		Action:     "create",
		Data:       `{"title": "Test Bookmark", "url": "https://example.com"}`,
		DeviceID:   "other-device",
		Timestamp:  time.Now().Add(-30 * time.Minute),
	}
	err := suite.db.Create(event).Error
	suite.Require().NoError(err)

	// Send sync request
	syncRequestMsg := &websocket.SyncMessage{
		Type:     "sync_request",
		UserID:   "test-user-123",
		DeviceID: "device-456",
		Data: map[string]interface{}{
			"last_sync_time": time.Now().Add(-1 * time.Hour).Unix(),
		},
		Timestamp: time.Now(),
	}

	response, err := suite.service.HandleSyncMessage(context.Background(), syncRequestMsg)
	suite.NoError(err)
	suite.Equal("sync_response", response.Type)
	suite.NotNil(response.Data)

	// Verify response contains events
	events := response.Data["events"].([]interface{})
	suite.Len(events, 1)
}

// Test real-time sync event creation and publishing
func (suite *WebSocketSyncIntegrationTestSuite) TestRealTimeSyncEventCreation() {
	// Mock Redis publish for broadcasting
	suite.redisClient.On("PublishSyncEvent", mock.Anything, "test-user-123", mock.AnythingOfType("*sync.SyncEvent")).Return(nil)

	// Create and publish a sync event
	event := &SyncEvent{
		Type:       SyncEventBookmarkUpdated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-123",
		Action:     "update",
		Data:       `{"title": "Updated Bookmark"}`,
		DeviceID:   "other-device",
		Timestamp:  time.Now(),
	}

	err := suite.service.CreateSyncEvent(context.Background(), event)
	suite.NoError(err)

	// Verify event was stored in database
	var storedEvent SyncEvent
	err = suite.db.Where("user_id = ? AND resource_id = ?", "test-user-123", "bookmark-123").First(&storedEvent).Error
	suite.NoError(err)
	suite.Equal(SyncEventBookmarkUpdated, storedEvent.Type)

	suite.redisClient.AssertExpected(suite.T())
}

// Test conflict resolution for WebSocket sync
func (suite *WebSocketSyncIntegrationTestSuite) TestConflictResolutionForWebSocket() {
	// Create conflicting events
	event1 := &SyncEvent{
		Type:       SyncEventBookmarkUpdated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-123",
		Action:     "update",
		Data:       `{"title": "Version 1"}`,
		DeviceID:   "device-1",
		Timestamp:  time.Now().Add(-2 * time.Minute),
	}

	event2 := &SyncEvent{
		Type:       SyncEventBookmarkUpdated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-123",
		Action:     "update",
		Data:       `{"title": "Version 2"}`,
		DeviceID:   "device-2",
		Timestamp:  time.Now().Add(-1 * time.Minute),
	}

	err := suite.db.Create(event1).Error
	suite.Require().NoError(err)
	err = suite.db.Create(event2).Error
	suite.Require().NoError(err)

	// Test conflict resolution
	conflicts := []*SyncEvent{event1, event2}
	winner := suite.service.ResolveConflict(conflicts)

	suite.NotNil(winner)
	suite.Equal(event2.ID, winner.ID) // Newer timestamp should win

	var winnerData map[string]interface{}
	err = json.Unmarshal([]byte(winner.Data), &winnerData)
	suite.NoError(err)
	suite.Equal("Version 2", winnerData["title"])
}

// Test offline queue processing for WebSocket
func (suite *WebSocketSyncIntegrationTestSuite) TestOfflineQueueProcessingForWebSocket() {
	// Mock Redis publish for processing offline queue
	suite.redisClient.On("PublishSyncEvent", mock.Anything, "test-user-123", mock.AnythingOfType("*sync.SyncEvent")).Return(nil)

	// Create offline events
	offlineEvent := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     "test-user-123",
		ResourceID: "bookmark-offline",
		Action:     "create",
		Data:       `{"title": "Offline Bookmark"}`,
		DeviceID:   "device-456",
		Status:     SyncStatusPending,
		Timestamp:  time.Now(),
	}

	err := suite.service.QueueOfflineEvent(context.Background(), offlineEvent)
	suite.NoError(err)

	// Process offline queue
	err = suite.service.ProcessOfflineQueue(context.Background(), "test-user-123", "device-456")
	suite.NoError(err)

	// Verify event status was updated
	var processedEvent SyncEvent
	err = suite.db.Where("user_id = ? AND resource_id = ?", "test-user-123", "bookmark-offline").First(&processedEvent).Error
	suite.NoError(err)
	suite.Equal(SyncStatusSynced, processedEvent.Status)

	suite.redisClient.AssertExpected(suite.T())
}
