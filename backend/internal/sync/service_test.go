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

// SyncServiceTestSuite defines the test suite for sync service
type SyncServiceTestSuite struct {
	suite.Suite
	service     *Service
	db          *gorm.DB
	redisClient *MockRedisClient
	logger      *zap.Logger
}

func (suite *SyncServiceTestSuite) SetupTest() {
	// Setup test database
	db, err := database.SetupTestDB()
	suite.Require().NoError(err)
	suite.db = db

	// Setup mock Redis client
	suite.redisClient = &MockRedisClient{}

	// Setup logger
	suite.logger = zap.NewNop()

	// Create service
	suite.service = NewService(suite.db, suite.redisClient, suite.logger)
}

func (suite *SyncServiceTestSuite) TearDownTest() {
	database.CleanupTestDB(suite.db)
}

func TestSyncServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SyncServiceTestSuite))
}

// Test sync event creation and publishing
func (suite *SyncServiceTestSuite) TestCreateSyncEvent() {
	userID := "test-user-123"
	bookmarkID := "bookmark-456"

	dataJSON, _ := json.Marshal(map[string]interface{}{
		"title": "Test Bookmark",
		"url":   "https://example.com",
	})

	event := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     userID,
		ResourceID: bookmarkID,
		Action:     "create",
		Data:       string(dataJSON),
		DeviceID:   "device-789",
		Timestamp:  time.Now(),
	}

	// Mock Redis publish
	suite.redisClient.On("PublishSyncEvent", mock.Anything, userID, mock.AnythingOfType("*sync.SyncEvent")).Return(nil)

	// Test creating sync event
	err := suite.service.CreateSyncEvent(context.Background(), event)
	suite.NoError(err)

	// Verify event was stored in database
	var storedEvent SyncEvent
	err = suite.db.Where("user_id = ? AND resource_id = ?", userID, bookmarkID).First(&storedEvent).Error
	suite.NoError(err)
	suite.Equal(event.Type, storedEvent.Type)
	suite.Equal(event.UserID, storedEvent.UserID)
	suite.Equal(event.ResourceID, storedEvent.ResourceID)
	suite.Equal(event.Action, storedEvent.Action)
	suite.Equal(event.DeviceID, storedEvent.DeviceID)

	// Verify Redis publish was called
	suite.redisClient.AssertExpected(suite.T())
}

// Test conflict resolution using timestamps
func (suite *SyncServiceTestSuite) TestResolveConflict() {
	userID := "test-user-123"
	bookmarkID := "bookmark-456"

	// Create two conflicting events
	data1JSON, _ := json.Marshal(map[string]interface{}{
		"title": "Old Title",
	})
	data2JSON, _ := json.Marshal(map[string]interface{}{
		"title": "New Title",
	})

	event1 := &SyncEvent{
		Type:       SyncEventBookmarkUpdated,
		UserID:     userID,
		ResourceID: bookmarkID,
		Action:     "update",
		Data:       string(data1JSON),
		DeviceID:   "device-1",
		Timestamp:  time.Now().Add(-1 * time.Minute), // Older timestamp
	}

	event2 := &SyncEvent{
		Type:       SyncEventBookmarkUpdated,
		UserID:     userID,
		ResourceID: bookmarkID,
		Action:     "update",
		Data:       string(data2JSON),
		DeviceID:   "device-2",
		Timestamp:  time.Now(), // Newer timestamp
	}

	// Test conflict resolution - newer timestamp should win
	winner := suite.service.ResolveConflict([]*SyncEvent{event1, event2})
	suite.Equal(event2, winner)

	// Parse the JSON data to verify content
	var winnerData map[string]interface{}
	err := json.Unmarshal([]byte(winner.Data), &winnerData)
	suite.NoError(err)
	suite.Equal("New Title", winnerData["title"])
}

// Test sync state tracking
func (suite *SyncServiceTestSuite) TestGetSyncState() {
	userID := "test-user-123"
	deviceID := "device-456"

	// Create some sync events
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       "{}",
			DeviceID:   deviceID,
			Timestamp:  time.Now().Add(-2 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       "{}",
			DeviceID:   deviceID,
			Timestamp:  time.Now().Add(-1 * time.Minute),
		},
	}

	// Store events in database
	for _, event := range events {
		err := suite.db.Create(event).Error
		suite.NoError(err)
	}

	// Test getting sync state
	state, err := suite.service.GetSyncState(context.Background(), userID, deviceID)
	suite.NoError(err)
	suite.NotNil(state)
	suite.Equal(userID, state.UserID)
	suite.Equal(deviceID, state.DeviceID)
	suite.True(state.LastSyncTime.After(time.Now().Add(-3 * time.Minute)))
}

// Test delta synchronization
func (suite *SyncServiceTestSuite) TestGetDeltaSync() {
	userID := "test-user-123"
	deviceID := "device-456"
	lastSyncTime := time.Now().Add(-1 * time.Hour)

	// Create events after last sync time
	recentEvents := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-new",
			Action:     "create",
			Data:       "{}",
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-updated",
			Action:     "update",
			Data:       "{}",
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-15 * time.Minute),
		},
	}

	// Create events before last sync time (should not be included)
	oldEvents := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-old",
			Action:     "create",
			Data:       "{}",
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-2 * time.Hour),
		},
	}

	// Store all events
	allEvents := append(recentEvents, oldEvents...)
	for _, event := range allEvents {
		err := suite.db.Create(event).Error
		suite.NoError(err)
	}

	// Test delta sync
	delta, err := suite.service.GetDeltaSync(context.Background(), userID, deviceID, lastSyncTime)
	suite.NoError(err)
	suite.Len(delta.Events, 2) // Only recent events
	suite.Equal("bookmark-new", delta.Events[0].ResourceID)
	suite.Equal("bookmark-updated", delta.Events[1].ResourceID)
}

// Test offline queue management
func (suite *SyncServiceTestSuite) TestOfflineQueue() {
	userID := "test-user-123"
	deviceID := "device-456"

	// Create offline events
	offlineEvents := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-offline-1",
			Action:     "create",
			Data:       "{}",
			DeviceID:   deviceID,
			Status:     SyncStatusPending,
			Timestamp:  time.Now().Add(-10 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-offline-2",
			Action:     "update",
			Data:       "{}",
			DeviceID:   deviceID,
			Status:     SyncStatusPending,
			Timestamp:  time.Now().Add(-5 * time.Minute),
		},
	}

	// Queue offline events
	for _, event := range offlineEvents {
		err := suite.service.QueueOfflineEvent(context.Background(), event)
		suite.NoError(err)
	}

	// Test getting offline queue
	queue, err := suite.service.GetOfflineQueue(context.Background(), userID, deviceID)
	suite.NoError(err)
	suite.Len(queue, 2)
	suite.Equal(SyncStatusPending, queue[0].Status)
	suite.Equal(SyncStatusPending, queue[1].Status)

	// Mock Redis publish for processing queue
	suite.redisClient.On("PublishSyncEvent", mock.Anything, userID, mock.AnythingOfType("*sync.SyncEvent")).Return(nil).Times(2)

	// Test processing offline queue
	err = suite.service.ProcessOfflineQueue(context.Background(), userID, deviceID)
	suite.NoError(err)

	// Verify events are marked as synced
	var processedEvents []SyncEvent
	err = suite.db.Where("user_id = ? AND device_id = ? AND status = ?", userID, deviceID, SyncStatusSynced).Find(&processedEvents).Error
	suite.NoError(err)
	suite.Len(processedEvents, 2)

	suite.redisClient.AssertExpected(suite.T())
}

// Test sync protocol message handling
func (suite *SyncServiceTestSuite) TestHandleSyncMessage() {
	userID := "test-user-123"
	deviceID := "device-456"

	// Test ping message
	pingMsg := &websocket.SyncMessage{
		Type:      "ping",
		UserID:    userID,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
	}

	response, err := suite.service.HandleSyncMessage(context.Background(), pingMsg)
	suite.NoError(err)
	suite.Equal("pong", response.Type)

	// Test sync request message
	syncRequestMsg := &websocket.SyncMessage{
		Type:     "sync_request",
		UserID:   userID,
		DeviceID: deviceID,
		Data: map[string]interface{}{
			"last_sync_time": time.Now().Add(-1 * time.Hour).Unix(),
		},
		Timestamp: time.Now(),
	}

	response, err = suite.service.HandleSyncMessage(context.Background(), syncRequestMsg)
	suite.NoError(err)
	suite.Equal("sync_response", response.Type)
	suite.NotNil(response.Data)
}

// Test bandwidth optimization
func (suite *SyncServiceTestSuite) TestBandwidthOptimization() {
	userID := "test-user-123"

	// Create multiple events for the same resource
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       "{}",
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       "{}",
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-20 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       "{}",
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-10 * time.Minute),
		},
	}

	// Test optimization - should return only the latest event per resource
	optimized := suite.service.OptimizeEvents(events)
	suite.Len(optimized, 1) // Only latest event should remain
	suite.Equal(SyncEventBookmarkUpdated, optimized[0].Type)
	suite.True(optimized[0].Timestamp.After(time.Now().Add(-15 * time.Minute)))
}
