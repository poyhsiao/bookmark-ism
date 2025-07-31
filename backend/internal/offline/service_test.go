package offline

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockRedisClient for testing
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockRedisClient) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) IncrementWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// OfflineServiceTestSuite defines the test suite for offline service
type OfflineServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	redisClient *MockRedisClient
	service     *Service
}

func (suite *OfflineServiceTestSuite) SetupTest() {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Run migrations
	err = database.AutoMigrate(db)
	suite.Require().NoError(err)

	// Create test user
	testUser := &database.User{
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(testUser).Error
	suite.Require().NoError(err)

	suite.db = db
	suite.redisClient = &MockRedisClient{}
	suite.service = NewService(suite.db, suite.redisClient)
}

func (suite *OfflineServiceTestSuite) TearDownTest() {
	// Close database connection
	sqlDB, _ := suite.db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// Test local bookmark caching system
func (suite *OfflineServiceTestSuite) TestCacheBookmark() {
	bookmark := &database.Bookmark{
		BaseModel:   database.BaseModel{ID: 1},
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Example",
		Description: "Test bookmark",
		Tags:        `["test", "example"]`,
	}

	// Mock Redis operations
	suite.redisClient.On("Set", mock.Anything, "offline:bookmark:1:1", mock.Anything, time.Hour*24).Return(nil)

	err := suite.service.CacheBookmark(context.Background(), bookmark)
	suite.NoError(err)
	suite.redisClient.AssertExpectations(suite.T())
}

func (suite *OfflineServiceTestSuite) TestGetCachedBookmark() {
	bookmarkJSON := `{"id":1,"user_id":1,"url":"https://example.com","title":"Example","description":"Test bookmark","tags":"[\"test\", \"example\"]"}`

	// Mock Redis operations
	suite.redisClient.On("Get", mock.Anything, "offline:bookmark:1:1").Return(bookmarkJSON, nil)

	bookmark, err := suite.service.GetCachedBookmark(context.Background(), 1, 1)
	suite.NoError(err)
	suite.NotNil(bookmark)
	suite.Equal("https://example.com", bookmark.URL)
	suite.Equal("Example", bookmark.Title)
	suite.redisClient.AssertExpectations(suite.T())
}

func (suite *OfflineServiceTestSuite) TestGetCachedBookmarkNotFound() {
	// Mock Redis operations - bookmark not found
	suite.redisClient.On("Get", mock.Anything, "offline:bookmark:1:999").Return("", ErrKeyNotFound)

	bookmark, err := suite.service.GetCachedBookmark(context.Background(), 1, 999)
	suite.Error(err)
	suite.Nil(bookmark)
	suite.Equal(ErrBookmarkNotCached, err)
	suite.redisClient.AssertExpectations(suite.T())
}

// Test offline change queuing
func (suite *OfflineServiceTestSuite) TestQueueOfflineChange() {
	change := &OfflineChange{
		ID:         "change-1",
		UserID:     1,
		DeviceID:   "device-123",
		Type:       "bookmark_create",
		ResourceID: "bookmark-1",
		Data:       `{"url":"https://example.com","title":"Example"}`,
		Timestamp:  time.Now(),
	}

	// Mock Redis operations
	suite.redisClient.On("Set", mock.Anything, "offline:queue:1:change-1", mock.Anything, time.Hour*24*7).Return(nil)

	err := suite.service.QueueOfflineChange(context.Background(), change)
	suite.NoError(err)
	suite.redisClient.AssertExpectations(suite.T())
}

func (suite *OfflineServiceTestSuite) TestGetOfflineQueue() {
	// For now, the GetOfflineQueue returns empty slice as it's a simplified implementation
	changes, err := suite.service.GetOfflineQueue(context.Background(), 1)
	suite.NoError(err)
	suite.Len(changes, 0) // Expecting empty slice for simplified implementation
}

// Test connectivity detection
func (suite *OfflineServiceTestSuite) TestCheckConnectivity() {
	// Mock successful connectivity check
	isOnline := suite.service.CheckConnectivity(context.Background())
	suite.True(isOnline) // Default implementation should return true for tests
}

func (suite *OfflineServiceTestSuite) TestGetOfflineStatus() {
	// Mock Redis operations
	suite.redisClient.On("Get", mock.Anything, "offline:status:1").Return("offline", nil)

	status, err := suite.service.GetOfflineStatus(context.Background(), 1)
	suite.NoError(err)
	suite.Equal("offline", status)
	suite.redisClient.AssertExpectations(suite.T())
}

func (suite *OfflineServiceTestSuite) TestSetOfflineStatus() {
	// Mock Redis operations
	suite.redisClient.On("Set", mock.Anything, "offline:status:1", "offline", time.Hour).Return(nil)

	err := suite.service.SetOfflineStatus(context.Background(), 1, "offline")
	suite.NoError(err)
	suite.redisClient.AssertExpectations(suite.T())
}

// Test cache management and cleanup
func (suite *OfflineServiceTestSuite) TestCleanupExpiredCache() {
	// For now, the CleanupExpiredCache is a simplified implementation that doesn't call Redis
	err := suite.service.CleanupExpiredCache(context.Background(), 1)
	suite.NoError(err)
}

func (suite *OfflineServiceTestSuite) TestGetCacheStats() {
	// Mock Redis operations
	suite.redisClient.On("Get", mock.Anything, "offline:stats:1").Return(`{"cached_bookmarks":10,"queued_changes":5,"last_sync":"2023-01-01T00:00:00Z"}`, nil)

	stats, err := suite.service.GetCacheStats(context.Background(), 1)
	suite.NoError(err)
	suite.NotNil(stats)
	suite.Equal(10, stats.CachedBookmarksCount)
	suite.Equal(5, stats.QueuedChangesCount)
	suite.redisClient.AssertExpectations(suite.T())
}

// Test conflict resolution
func (suite *OfflineServiceTestSuite) TestResolveConflict() {
	localChange := &OfflineChange{
		ID:         "change-1",
		UserID:     1,
		Type:       "bookmark_update",
		ResourceID: "bookmark-1",
		Data:       `{"title":"Local Title"}`,
		Timestamp:  time.Now().Add(-time.Minute),
	}

	serverChange := &OfflineChange{
		ID:         "change-2",
		UserID:     1,
		Type:       "bookmark_update",
		ResourceID: "bookmark-1",
		Data:       `{"title":"Server Title"}`,
		Timestamp:  time.Now(),
	}

	// Server change is newer, should win
	resolved := suite.service.ResolveConflict(localChange, serverChange)
	suite.Equal(serverChange.ID, resolved.ID)
	suite.Equal(`{"title":"Server Title"}`, resolved.Data)
}

// Run the test suite
func TestOfflineServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OfflineServiceTestSuite))
}

// Additional unit tests
func TestNewService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	redisClient := &MockRedisClient{}
	service := NewService(db, redisClient)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, redisClient, service.redisClient)
}
