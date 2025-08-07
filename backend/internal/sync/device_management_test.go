package sync

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeviceManagementTestSuite tests device registration and identification
type DeviceManagementTestSuite struct {
	suite.Suite
	service     *Service
	db          *gorm.DB
	redisClient *MockRedisClient
	logger      *zap.Logger
}

func (suite *DeviceManagementTestSuite) SetupTest() {
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

func (suite *DeviceManagementTestSuite) TearDownTest() {
	database.CleanupTestDB(suite.db)
}

func TestDeviceManagementTestSuite(t *testing.T) {
	suite.Run(t, new(DeviceManagementTestSuite))
}

// Test device registration and sync state creation
func (suite *DeviceManagementTestSuite) TestDeviceRegistration() {
	userID := "test-user-123"
	deviceID := "device-456"

	// First access should create sync state
	state, err := suite.service.GetSyncState(context.Background(), userID, deviceID)
	suite.NoError(err)
	suite.NotNil(state)
	suite.Equal(userID, state.UserID)
	suite.Equal(deviceID, state.DeviceID)
	suite.True(state.LastSyncTime.After(time.Now().Add(-1 * time.Minute)))

	// Verify state was persisted in database
	var dbState SyncState
	err = suite.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&dbState).Error
	suite.NoError(err)
	suite.Equal(state.ID, dbState.ID)
}

// Test multiple device registration for same user
func (suite *DeviceManagementTestSuite) TestMultipleDeviceRegistration() {
	userID := "test-user-123"
	device1ID := "device-1"
	device2ID := "device-2"

	// Register first device
	state1, err := suite.service.GetSyncState(context.Background(), userID, device1ID)
	suite.NoError(err)
	suite.Equal(device1ID, state1.DeviceID)

	// Register second device
	state2, err := suite.service.GetSyncState(context.Background(), userID, device2ID)
	suite.NoError(err)
	suite.Equal(device2ID, state2.DeviceID)

	// Verify both devices are registered
	var states []SyncState
	err = suite.db.Where("user_id = ?", userID).Find(&states).Error
	suite.NoError(err)
	suite.Len(states, 2)

	deviceIDs := make([]string, len(states))
	for i, state := range states {
		deviceIDs[i] = state.DeviceID
	}
	suite.Contains(deviceIDs, device1ID)
	suite.Contains(deviceIDs, device2ID)
}

// Test sync state update
func (suite *DeviceManagementTestSuite) TestSyncStateUpdate() {
	userID := "test-user-123"
	deviceID := "device-456"

	// Create initial sync state
	initialState, err := suite.service.GetSyncState(context.Background(), userID, deviceID)
	suite.NoError(err)
	initialTime := initialState.LastSyncTime

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update sync state
	newSyncTime := time.Now()
	err = suite.service.UpdateSyncState(context.Background(), userID, deviceID, newSyncTime)
	suite.NoError(err)

	// Verify state was updated
	updatedState, err := suite.service.GetSyncState(context.Background(), userID, deviceID)
	suite.NoError(err)
	suite.True(updatedState.LastSyncTime.After(initialTime))
	suite.True(updatedState.LastSyncTime.Sub(newSyncTime).Abs() < time.Second)
}

// Test sync state for non-existent device
func (suite *DeviceManagementTestSuite) TestSyncStateCreationForNewDevice() {
	userID := "test-user-123"
	deviceID := "new-device-789"

	// Getting sync state for non-existent device should create it
	state, err := suite.service.GetSyncState(context.Background(), userID, deviceID)
	suite.NoError(err)
	suite.NotNil(state)
	suite.Equal(userID, state.UserID)
	suite.Equal(deviceID, state.DeviceID)
	suite.NotZero(state.ID)

	// Verify it was created in database
	var dbState SyncState
	err = suite.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&dbState).Error
	suite.NoError(err)
	suite.Equal(state.ID, dbState.ID)
}

// Test device identification in sync events
func (suite *DeviceManagementTestSuite) TestDeviceIdentificationInSyncEvents() {
	userID := "test-user-123"
	device1ID := "device-1"
	device2ID := "device-2"

	// Mock Redis publish
	suite.redisClient.On("PublishSyncEvent", context.Background(), userID, mock.AnythingOfType("*sync.SyncEvent")).Return(nil).Times(2)

	// Create sync events from different devices
	event1 := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     userID,
		ResourceID: "bookmark-1",
		Action:     "create",
		Data:       `{"title": "Bookmark from Device 1"}`,
		DeviceID:   device1ID,
		Timestamp:  time.Now(),
	}

	event2 := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     userID,
		ResourceID: "bookmark-2",
		Action:     "create",
		Data:       `{"title": "Bookmark from Device 2"}`,
		DeviceID:   device2ID,
		Timestamp:  time.Now(),
	}

	err := suite.service.CreateSyncEvent(context.Background(), event1)
	suite.NoError(err)

	err = suite.service.CreateSyncEvent(context.Background(), event2)
	suite.NoError(err)

	// Verify events are stored with correct device IDs
	var events []SyncEvent
	err = suite.db.Where("user_id = ?", userID).Find(&events).Error
	suite.NoError(err)
	suite.Len(events, 2)

	deviceIDs := make([]string, len(events))
	for i, event := range events {
		deviceIDs[i] = event.DeviceID
	}
	suite.Contains(deviceIDs, device1ID)
	suite.Contains(deviceIDs, device2ID)

	suite.redisClient.AssertExpected(suite.T())
}

// Test delta sync excludes events from same device
func (suite *DeviceManagementTestSuite) TestDeltaSyncDeviceExclusion() {
	userID := "test-user-123"
	deviceID := "device-456"
	otherDeviceID := "other-device-789"

	// Create events from same device (should be excluded)
	sameDeviceEvent := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     userID,
		ResourceID: "bookmark-same",
		Action:     "create",
		Data:       `{"title": "Same Device Bookmark"}`,
		DeviceID:   deviceID,
		Timestamp:  time.Now().Add(-30 * time.Minute),
	}

	// Create events from other device (should be included)
	otherDeviceEvent := &SyncEvent{
		Type:       SyncEventBookmarkCreated,
		UserID:     userID,
		ResourceID: "bookmark-other",
		Action:     "create",
		Data:       `{"title": "Other Device Bookmark"}`,
		DeviceID:   otherDeviceID,
		Timestamp:  time.Now().Add(-30 * time.Minute),
	}

	err := suite.db.Create(sameDeviceEvent).Error
	suite.Require().NoError(err)
	err = suite.db.Create(otherDeviceEvent).Error
	suite.Require().NoError(err)

	// Get delta sync for the device
	lastSyncTime := time.Now().Add(-1 * time.Hour)
	delta, err := suite.service.GetDeltaSync(context.Background(), userID, deviceID, lastSyncTime)
	suite.NoError(err)

	// Should only include events from other devices
	suite.Len(delta.Events, 1)
	suite.Equal(otherDeviceID, delta.Events[0].DeviceID)
	suite.Equal("bookmark-other", delta.Events[0].ResourceID)
}

// Test sync history tracking
func (suite *DeviceManagementTestSuite) TestSyncHistoryTracking() {
	userID := "test-user-123"
	deviceID := "device-456"

	// Create multiple sync events over time
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "First Bookmark"}`,
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-3 * time.Hour),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Updated Bookmark"}`,
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-2 * time.Hour),
		},
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-2",
			Action:     "create",
			Data:       `{"title": "Second Bookmark"}`,
			DeviceID:   "other-device",
			Timestamp:  time.Now().Add(-1 * time.Hour),
		},
	}

	for _, event := range events {
		err := suite.db.Create(event).Error
		suite.Require().NoError(err)
	}

	// Test delta sync with different time ranges

	// Get all events (last 4 hours)
	// Note: OptimizeEvents will merge multiple events for the same resource, so we expect 2 events (bookmark-1 latest + bookmark-2)
	delta1, err := suite.service.GetDeltaSync(context.Background(), userID, deviceID, time.Now().Add(-4*time.Hour))
	suite.NoError(err)
	suite.Len(delta1.Events, 2) // bookmark-1 (latest update) + bookmark-2

	// Get recent events (last 90 minutes)
	delta2, err := suite.service.GetDeltaSync(context.Background(), userID, deviceID, time.Now().Add(-90*time.Minute))
	suite.NoError(err)
	suite.Len(delta2.Events, 1)
	suite.Equal("bookmark-2", delta2.Events[0].ResourceID)

	// Get no events (last 30 minutes)
	delta3, err := suite.service.GetDeltaSync(context.Background(), userID, deviceID, time.Now().Add(-30*time.Minute))
	suite.NoError(err)
	suite.Len(delta3.Events, 0)
}
