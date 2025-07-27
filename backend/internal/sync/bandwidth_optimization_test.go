package sync

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BandwidthOptimizationTestSuite tests bandwidth optimization features
type BandwidthOptimizationTestSuite struct {
	suite.Suite
	service     *Service
	db          *gorm.DB
	redisClient *MockRedisClient
	logger      *zap.Logger
}

func (suite *BandwidthOptimizationTestSuite) SetupTest() {
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

func (suite *BandwidthOptimizationTestSuite) TearDownTest() {
	database.CleanupTestDB(suite.db)
}

func TestBandwidthOptimizationTestSuite(t *testing.T) {
	suite.Run(t, new(BandwidthOptimizationTestSuite))
}

// Test event optimization for same resource
func (suite *BandwidthOptimizationTestSuite) TestEventOptimizationSameResource() {
	userID := "test-user-123"

	// Create multiple events for the same resource
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "Original Title"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Updated Title"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-20 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Final Title"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-10 * time.Minute),
		},
	}

	// Test optimization - should return only the latest event per resource
	optimized := suite.service.OptimizeEvents(events)
	suite.Len(optimized, 1) // Only latest event should remain
	suite.Equal(SyncEventBookmarkUpdated, optimized[0].Type)
	suite.Contains(optimized[0].Data, "Final Title")
	suite.True(optimized[0].Timestamp.After(time.Now().Add(-15 * time.Minute)))
}

// Test event optimization for multiple resources
func (suite *BandwidthOptimizationTestSuite) TestEventOptimizationMultipleResources() {
	userID := "test-user-123"

	// Create events for different resources
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "Bookmark 1"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Updated Bookmark 1"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-20 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-2",
			Action:     "create",
			Data:       `{"title": "Bookmark 2"}`,
			DeviceID:   "device-2",
			Timestamp:  time.Now().Add(-25 * time.Minute),
		},
		{
			Type:       SyncEventCollectionCreated,
			UserID:     userID,
			ResourceID: "collection-1",
			Action:     "create",
			Data:       `{"name": "My Collection"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-15 * time.Minute),
		},
	}

	// Test optimization - should return latest event per resource
	optimized := suite.service.OptimizeEvents(events)
	suite.Len(optimized, 3) // bookmark-1 (latest), bookmark-2, collection-1

	// Verify events are sorted by timestamp
	for i := 1; i < len(optimized); i++ {
		suite.True(optimized[i].Timestamp.After(optimized[i-1].Timestamp) ||
			optimized[i].Timestamp.Equal(optimized[i-1].Timestamp))
	}

	// Verify we have the latest version of bookmark-1
	bookmark1Event := findEventByResourceID(optimized, "bookmark-1")
	suite.NotNil(bookmark1Event)
	suite.Equal(SyncEventBookmarkUpdated, bookmark1Event.Type)
	suite.Contains(bookmark1Event.Data, "Updated Bookmark 1")
}

// Test delta sync with bandwidth optimization
func (suite *BandwidthOptimizationTestSuite) TestDeltaSyncWithOptimization() {
	userID := "test-user-123"
	deviceID := "device-456"

	// Create multiple events for same resource from different devices
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "Original"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-60 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Updated"}`,
			DeviceID:   "device-2",
			Timestamp:  time.Now().Add(-45 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Final"}`,
			DeviceID:   "device-3",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-2",
			Action:     "create",
			Data:       `{"title": "Another Bookmark"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-40 * time.Minute),
		},
	}

	// Store events in database
	for _, event := range events {
		err := suite.db.Create(event).Error
		suite.Require().NoError(err)
	}

	// Get delta sync (should be optimized)
	lastSyncTime := time.Now().Add(-2 * time.Hour)
	delta, err := suite.service.GetDeltaSync(context.Background(), userID, deviceID, lastSyncTime)
	suite.NoError(err)

	// Should return optimized events (latest per resource)
	suite.Len(delta.Events, 2) // bookmark-1 (latest) + bookmark-2

	// Verify we got the latest version of bookmark-1
	bookmark1Event := findEventByResourceID(delta.Events, "bookmark-1")
	suite.NotNil(bookmark1Event)
	suite.Contains(bookmark1Event.Data, "Final")
	suite.Equal("device-3", bookmark1Event.DeviceID)
}

// Test optimization with delete events
func (suite *BandwidthOptimizationTestSuite) TestOptimizationWithDeleteEvents() {
	userID := "test-user-123"

	// Create events including delete
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "Created"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkUpdated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "update",
			Data:       `{"title": "Updated"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-20 * time.Minute),
		},
		{
			Type:       SyncEventBookmarkDeleted,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "delete",
			Data:       `{}`,
			DeviceID:   "device-2",
			Timestamp:  time.Now().Add(-10 * time.Minute),
		},
	}

	// Test optimization - should return only the delete event (latest)
	optimized := suite.service.OptimizeEvents(events)
	suite.Len(optimized, 1)
	suite.Equal(SyncEventBookmarkDeleted, optimized[0].Type)
	suite.Equal("delete", optimized[0].Action)
}

// Test optimization preserves chronological order
func (suite *BandwidthOptimizationTestSuite) TestOptimizationPreservesOrder() {
	userID := "test-user-123"

	// Create events for different resources at different times
	events := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-3",
			Action:     "create",
			Data:       `{"title": "Third"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-10 * time.Minute), // Latest
		},
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "First"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-30 * time.Minute), // Earliest
		},
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     userID,
			ResourceID: "bookmark-2",
			Action:     "create",
			Data:       `{"title": "Second"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now().Add(-20 * time.Minute), // Middle
		},
	}

	// Test optimization preserves chronological order
	optimized := suite.service.OptimizeEvents(events)
	suite.Len(optimized, 3)

	// Should be ordered by timestamp (earliest first)
	suite.Equal("bookmark-1", optimized[0].ResourceID)
	suite.Equal("bookmark-2", optimized[1].ResourceID)
	suite.Equal("bookmark-3", optimized[2].ResourceID)
}

// Test empty events optimization
func (suite *BandwidthOptimizationTestSuite) TestEmptyEventsOptimization() {
	// Test with empty slice
	optimized := suite.service.OptimizeEvents([]*SyncEvent{})
	suite.Len(optimized, 0)

	// Test with single event
	singleEvent := []*SyncEvent{
		{
			Type:       SyncEventBookmarkCreated,
			UserID:     "test-user-123",
			ResourceID: "bookmark-1",
			Action:     "create",
			Data:       `{"title": "Single"}`,
			DeviceID:   "device-1",
			Timestamp:  time.Now(),
		},
	}

	optimized = suite.service.OptimizeEvents(singleEvent)
	suite.Len(optimized, 1)
	suite.Equal(singleEvent[0], optimized[0])
}

// Helper function to find event by resource ID
func findEventByResourceID(events []*SyncEvent, resourceID string) *SyncEvent {
	for _, event := range events {
		if event.ResourceID == resourceID {
			return event
		}
	}
	return nil
}
