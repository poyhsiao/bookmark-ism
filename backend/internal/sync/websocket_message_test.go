package sync

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/websocket"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebSocketMessageTestSuite tests the sync service integration with WebSocket messages
type WebSocketMessageTestSuite struct {
	suite.Suite
	service     *Service
	db          *gorm.DB
	redisClient *MockRedisClient
	logger      *zap.Logger
}

func (suite *WebSocketMessageTestSuite) SetupTest() {
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

func (suite *WebSocketMessageTestSuite) TearDownTest() {
	database.CleanupTestDB(suite.db)
}

func TestWebSocketMessageTestSuite(t *testing.T) {
	suite.Run(t, new(WebSocketMessageTestSuite))
}

// Test WebSocket ping message handling
func (suite *WebSocketMessageTestSuite) TestWebSocketPingMessage() {
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

// Test WebSocket sync request message handling
func (suite *WebSocketMessageTestSuite) TestWebSocketSyncRequestMessage() {
	// Create some test sync events first
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

// Test WebSocket sync message with invalid type
func (suite *WebSocketMessageTestSuite) TestWebSocketInvalidMessage() {
	// Test unknown message type
	unknownMsg := &websocket.SyncMessage{
		Type:      "unknown_type",
		UserID:    "test-user-123",
		DeviceID:  "device-456",
		Data:      map[string]interface{}{"test": "data"},
		Timestamp: time.Now(),
	}

	response, err := suite.service.HandleSyncMessage(context.Background(), unknownMsg)
	suite.Error(err)
	suite.Nil(response)
	suite.Contains(err.Error(), "unknown message type")
}

// Test WebSocket sync request with empty data
func (suite *WebSocketMessageTestSuite) TestWebSocketSyncRequestEmptyData() {
	// Send sync request with no data
	syncRequestMsg := &websocket.SyncMessage{
		Type:      "sync_request",
		UserID:    "test-user-123",
		DeviceID:  "device-456",
		Data:      nil,
		Timestamp: time.Now(),
	}

	response, err := suite.service.HandleSyncMessage(context.Background(), syncRequestMsg)
	suite.NoError(err)
	suite.Equal("sync_response", response.Type)
	suite.NotNil(response.Data)

	// Should return empty events list
	events := response.Data["events"].([]interface{})
	suite.Len(events, 0)
}

// Test WebSocket sync request with malformed last_sync_time
func (suite *WebSocketMessageTestSuite) TestWebSocketSyncRequestMalformedTime() {
	// Send sync request with invalid last_sync_time
	syncRequestMsg := &websocket.SyncMessage{
		Type:     "sync_request",
		UserID:   "test-user-123",
		DeviceID: "device-456",
		Data: map[string]interface{}{
			"last_sync_time": "invalid_time",
		},
		Timestamp: time.Now(),
	}

	response, err := suite.service.HandleSyncMessage(context.Background(), syncRequestMsg)
	suite.NoError(err)
	suite.Equal("sync_response", response.Type)
	suite.NotNil(response.Data)

	// Should default to 24 hours ago and return empty events
	events := response.Data["events"].([]interface{})
	suite.Len(events, 0)
}
