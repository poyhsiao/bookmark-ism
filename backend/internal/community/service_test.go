package community

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Mock interfaces
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Where(query interface{}, args ...interface{}) Database {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(Database)
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Order(value interface{}) Database {
	args := m.Called(value)
	return args.Get(0).(Database)
}

func (m *MockDB) Limit(limit int) Database {
	args := m.Called(limit)
	return args.Get(0).(Database)
}

func (m *MockDB) Offset(offset int) Database {
	args := m.Called(offset)
	return args.Get(0).(Database)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockRedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

// Test Suite
type CommunityServiceTestSuite struct {
	suite.Suite
	service   *Service
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *CommunityServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()

	suite.service = &Service{
		db:    suite.mockDB,
		redis: suite.mockRedis,
	}
}

func (suite *CommunityServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// Test User Behavior Tracking
func (suite *CommunityServiceTestSuite) TestTrackUserBehavior() {
	request := &BehaviorTrackingRequest{
		UserID:     "user-123",
		BookmarkID: 1,
		ActionType: "view",
		Duration:   30,
		Context:    "homepage",
		Metadata:   map[string]interface{}{"source": "recommendation"},
	}

	// Mock successful database creation
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserBehavior")).Return(&gorm.DB{Error: nil})

	err := suite.service.TrackUserBehavior(suite.ctx, request)

	assert.NoError(suite.T(), err)
}

func (suite *CommunityServiceTestSuite) TestTrackUserBehaviorInvalidRequest() {
	request := &BehaviorTrackingRequest{
		UserID:     "", // Invalid empty user ID
		BookmarkID: 1,
		ActionType: "view",
	}

	err := suite.service.TrackUserBehavior(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test User Following
func (suite *CommunityServiceTestSuite) TestFollowUser() {
	request := &FollowRequest{
		FollowingID: "user-456",
	}
	followerID := "user-123"

	// Mock check for existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	// Mock successful creation
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserFollow")).Return(&gorm.DB{Error: nil})

	// Mock Redis cache clearing calls (clearUserStatsCache is called for both users)
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-123"}).Return(nil)
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-456"}).Return(nil)

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.NoError(suite.T(), err)
}

func (suite *CommunityServiceTestSuite) TestFollowUserAlreadyFollowing() {
	request := &FollowRequest{
		FollowingID: "user-456",
	}
	followerID := "user-123"

	// Mock existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrAlreadyFollowing, err)
}

func (suite *CommunityServiceTestSuite) TestFollowUserSelf() {
	request := &FollowRequest{
		FollowingID: "user-123",
	}
	followerID := "user-123"

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrCannotFollowSelf, err)
}

// Test Unfollow User
func (suite *CommunityServiceTestSuite) TestUnfollowUser() {
	followingID := "user-456"
	followerID := "user-123"

	// Mock existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock successful deletion
	suite.mockDB.On("Delete", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock Redis cache clearing calls (clearUserStatsCache is called for both users)
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-123"}).Return(nil)
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-456"}).Return(nil)

	err := suite.service.UnfollowUser(suite.ctx, followerID, followingID)

	assert.NoError(suite.T(), err)
}

func (suite *CommunityServiceTestSuite) TestUnfollowUserNotFollowing() {
	followingID := "user-456"
	followerID := "user-123"

	// Mock no existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	err := suite.service.UnfollowUser(suite.ctx, followerID, followingID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrNotFollowing, err)
}

// Test Get Recommendations
func (suite *CommunityServiceTestSuite) TestGetRecommendations() {
	request := &RecommendationRequest{
		UserID:    "user-123",
		Limit:     10,
		Algorithm: "collaborative",
		Context:   "homepage",
	}

	// Mock cache miss (empty string returned)
	suite.mockRedis.On("Get", suite.ctx, "recommendations:user-123:collaborative:homepage").Return("", nil)

	// Mock finding existing recommendations
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.BookmarkRecommendation"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock cache set call
	suite.mockRedis.On("Set", suite.ctx, "recommendations:user-123:collaborative:homepage", mock.Anything, 15*time.Minute).Return(nil)

	recommendations, err := suite.service.GetRecommendations(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)
}

func (suite *CommunityServiceTestSuite) TestGetRecommendationsInvalidRequest() {
	request := &RecommendationRequest{
		UserID: "", // Invalid empty user ID
		Limit:  10,
	}

	recommendations, err := suite.service.GetRecommendations(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), recommendations)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test GetRecommendations cache hit - TDD: Write failing test first
func (suite *CommunityServiceTestSuite) TestGetRecommendationsCacheHit() {
	request := &RecommendationRequest{
		UserID:    "user-123",
		Limit:     10,
		Algorithm: "collaborative",
		Context:   "homepage",
	}

	// Expected cached recommendations
	expectedRecommendations := []RecommendationResponse{
		{
			BookmarkID: 1,
			Score:      0.9,
			ReasonType: "collaborative",
			ReasonText: "Users with similar interests also liked this",
		},
		{
			BookmarkID: 2,
			Score:      0.8,
			ReasonType: "collaborative",
			ReasonText: "Users with similar interests also liked this",
		},
	}

	// Mock cache hit - return cached recommendations
	cachedData := `[{"bookmark_id":1,"score":0.9,"reason_type":"collaborative","reason_text":"Users with similar interests also liked this"},{"bookmark_id":2,"score":0.8,"reason_type":"collaborative","reason_text":"Users with similar interests also liked this"}]`
	suite.mockRedis.On("Get", suite.ctx, "recommendations:user-123:collaborative:homepage").Return(cachedData, nil)

	recommendations, err := suite.service.GetRecommendations(suite.ctx, request)

	// Verify cache hit behavior
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)
	assert.Len(suite.T(), recommendations, 2)
	assert.Equal(suite.T(), expectedRecommendations[0].BookmarkID, recommendations[0].BookmarkID)
	assert.Equal(suite.T(), expectedRecommendations[0].Score, recommendations[0].Score)
	assert.Equal(suite.T(), expectedRecommendations[1].BookmarkID, recommendations[1].BookmarkID)
	assert.Equal(suite.T(), expectedRecommendations[1].Score, recommendations[1].Score)

	// Verify database was NOT called (cache hit)
	suite.mockDB.AssertNotCalled(suite.T(), "Where")
	suite.mockDB.AssertNotCalled(suite.T(), "Find")
}

// Test Get Trending Bookmarks
func (suite *CommunityServiceTestSuite) TestGetTrendingBookmarks() {
	request := &TrendingRequest{
		TimeWindow: "daily",
		Limit:      20,
		MinScore:   0.5,
	}

	// Mock finding trending bookmarks
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.TrendingBookmark"), mock.Anything).Return(&gorm.DB{Error: nil})

	trending, err := suite.service.GetTrendingBookmarksInternal(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), trending)
}

func (suite *CommunityServiceTestSuite) TestGetTrendingBookmarksInvalidTimeWindow() {
	request := &TrendingRequest{
		TimeWindow: "invalid", // Invalid time window
		Limit:      20,
	}

	trending, err := suite.service.GetTrendingBookmarksInternal(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), trending)
	assert.Equal(suite.T(), ErrInvalidTimeWindow, err)
}

// Test Generate User Feed
func (suite *CommunityServiceTestSuite) TestGenerateUserFeed() {
	request := &FeedRequest{
		UserID:     "user-123",
		Limit:      20,
		Offset:     0,
		SourceType: "all",
	}

	// Mock finding feed items
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Offset", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Return(&gorm.DB{Error: nil})

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), feed)
}

func (suite *CommunityServiceTestSuite) TestGenerateUserFeedInvalidRequest() {
	request := &FeedRequest{
		UserID: "", // Invalid empty user ID
		Limit:  20,
	}

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), feed)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test Get Social Metrics
func (suite *CommunityServiceTestSuite) TestGetSocialMetrics() {
	bookmarkID := uint(1)

	// Mock finding social metrics
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: nil})

	metrics, err := suite.service.GetSocialMetrics(suite.ctx, bookmarkID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), metrics)
}

func (suite *CommunityServiceTestSuite) TestGetSocialMetricsNotFound() {
	bookmarkID := uint(999)

	// Mock not finding social metrics
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	metrics, err := suite.service.GetSocialMetrics(suite.ctx, bookmarkID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), metrics)
	assert.Equal(suite.T(), ErrBookmarkNotFound, err)
}

// Test Update Social Metrics
func (suite *CommunityServiceTestSuite) TestUpdateSocialMetrics() {
	bookmarkID := uint(1)
	actionType := "view"

	// Mock finding existing metrics (record not found, so Create will be called)
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	// Mock creating new metrics
	suite.mockDB.On("Create", mock.AnythingOfType("*community.SocialMetrics")).Return(&gorm.DB{Error: nil})

	err := suite.service.UpdateSocialMetrics(suite.ctx, bookmarkID, actionType)

	assert.NoError(suite.T(), err)
}

func (suite *CommunityServiceTestSuite) TestUpdateSocialMetricsCreateNew() {
	bookmarkID := uint(1)
	actionType := "view"

	// Mock not finding existing metrics
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	// Mock creating new metrics
	suite.mockDB.On("Create", mock.AnythingOfType("*community.SocialMetrics")).Return(&gorm.DB{Error: nil})

	err := suite.service.UpdateSocialMetrics(suite.ctx, bookmarkID, actionType)

	assert.NoError(suite.T(), err)
}

// Test Get User Stats
func (suite *CommunityServiceTestSuite) TestGetUserStats() {
	userID := "user-123"

	// Mock cache miss (empty string returned)
	suite.mockRedis.On("Get", suite.ctx, "user_stats:user-123").Return("", nil)

	// Mock various database queries for user stats
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock cache set call
	suite.mockRedis.On("Set", suite.ctx, "user_stats:user-123", mock.Anything, 30*time.Minute).Return(nil)

	stats, err := suite.service.GetUserStats(suite.ctx, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), userID, stats.UserID)
}

func (suite *CommunityServiceTestSuite) TestGetUserStatsInvalidUserID() {
	userID := ""

	stats, err := suite.service.GetUserStats(suite.ctx, userID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), stats)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test Calculate Trending Scores
func (suite *CommunityServiceTestSuite) TestCalculateTrendingScores() {
	timeWindow := "daily"

	// Mock finding user behaviors - return empty slice (no behaviors to process)
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Return(&gorm.DB{Error: nil})

	// No database operations should be called for trending bookmarks since there are no behaviors
	err := suite.service.CalculateTrendingScores(suite.ctx, timeWindow)

	assert.NoError(suite.T(), err)
}

func (suite *CommunityServiceTestSuite) TestCalculateTrendingScoresInvalidTimeWindow() {
	timeWindow := "invalid"

	err := suite.service.CalculateTrendingScores(suite.ctx, timeWindow)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidTimeWindow, err)
}

// Test Generate Recommendations
func (suite *CommunityServiceTestSuite) TestGenerateRecommendations() {
	userID := "user-123"
	algorithm := "collaborative"

	// Mock finding user behaviors - return empty slice (no behaviors)
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Return(&gorm.DB{Error: nil})

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	// Should return insufficient data error when no behaviors exist
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInsufficientData, err)
}

func (suite *CommunityServiceTestSuite) TestGenerateRecommendationsInvalidAlgorithm() {
	userID := "user-123"
	algorithm := "invalid"

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidAlgorithm, err)
}

// Test GenerateRecommendations insufficient data - TDD: Write failing test first
func (suite *CommunityServiceTestSuite) TestGenerateRecommendationsInsufficientData() {
	userID := "user-123"
	algorithm := "collaborative"

	// Mock finding empty user behaviors (insufficient data)
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Return(&gorm.DB{Error: nil})

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	// Verify insufficient data error is returned
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInsufficientData, err)

	// Verify no recommendations are created when there's insufficient data
	suite.mockDB.AssertNotCalled(suite.T(), "Create", mock.AnythingOfType("*community.BookmarkRecommendation"))
}

// Run the test suite
func TestCommunityServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CommunityServiceTestSuite))
}

// Simple unit test for insufficient data - TDD approach
func TestGenerateRecommendationsInsufficientDataSimple(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis, nil, nil)

	ctx := context.Background()
	userID := "user-123"
	algorithm := "collaborative"

	// Mock finding empty user behaviors (insufficient data)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Run(func(args mock.Arguments) {
		// Return empty behaviors slice to simulate insufficient data
		behaviorsPtr := args.Get(0).(*[]UserBehavior)
		*behaviorsPtr = []UserBehavior{} // Empty slice
	}).Return(&gorm.DB{Error: nil})

	err := service.GenerateRecommendations(ctx, userID, algorithm)

	// Verify insufficient data error is returned
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientData, err)

	// Verify no recommendations are created when there's insufficient data
	mockDB.AssertNotCalled(t, "Create", mock.AnythingOfType("*community.BookmarkRecommendation"))
	mockDB.AssertExpectations(t)
}

// Test with sufficient data to ensure normal flow still works
func TestGenerateRecommendationsSufficientDataSimple(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis, nil, nil)

	ctx := context.Background()
	userID := "user-123"
	algorithm := "collaborative"

	// Mock finding user behaviors (sufficient data)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Run(func(args mock.Arguments) {
		// Return some behaviors to simulate sufficient data
		behaviorsPtr := args.Get(0).(*[]UserBehavior)
		*behaviorsPtr = []UserBehavior{
			{UserID: userID, BookmarkID: 1, ActionType: "view"},
			{UserID: userID, BookmarkID: 2, ActionType: "save"},
		}
	}).Return(&gorm.DB{Error: nil})

	// Mock creating recommendations
	mockDB.On("Create", mock.AnythingOfType("*community.BookmarkRecommendation")).Return(&gorm.DB{Error: nil})

	err := service.GenerateRecommendations(ctx, userID, algorithm)

	// Verify no error is returned when there's sufficient data
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

// Additional unit tests for validation methods
func TestUserBehaviorValidation(t *testing.T) {
	tests := []struct {
		name     string
		behavior UserBehavior
		wantErr  error
	}{
		{
			name: "Valid behavior",
			behavior: UserBehavior{
				UserID:     "user-123",
				BookmarkID: 1,
				ActionType: "view",
			},
			wantErr: nil,
		},
		{
			name: "Invalid user ID",
			behavior: UserBehavior{
				UserID:     "",
				BookmarkID: 1,
				ActionType: "view",
			},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "Invalid bookmark ID",
			behavior: UserBehavior{
				UserID:     "user-123",
				BookmarkID: 0,
				ActionType: "view",
			},
			wantErr: ErrInvalidBookmarkID,
		},
		{
			name: "Invalid action type",
			behavior: UserBehavior{
				UserID:     "user-123",
				BookmarkID: 1,
				ActionType: "invalid",
			},
			wantErr: ErrInvalidActionType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.behavior.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestUserFollowValidation(t *testing.T) {
	tests := []struct {
		name    string
		follow  UserFollow
		wantErr error
	}{
		{
			name: "Valid follow",
			follow: UserFollow{
				FollowerID:  "user-123",
				FollowingID: "user-456",
				Status:      "active",
			},
			wantErr: nil,
		},
		{
			name: "Cannot follow self",
			follow: UserFollow{
				FollowerID:  "user-123",
				FollowingID: "user-123",
				Status:      "active",
			},
			wantErr: ErrCannotFollowSelf,
		},
		{
			name: "Invalid status",
			follow: UserFollow{
				FollowerID:  "user-123",
				FollowingID: "user-456",
				Status:      "invalid",
			},
			wantErr: ErrInvalidFollowStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.follow.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestBookmarkRecommendationValidation(t *testing.T) {
	tests := []struct {
		name           string
		recommendation BookmarkRecommendation
		wantErr        error
	}{
		{
			name: "Valid recommendation",
			recommendation: BookmarkRecommendation{
				UserID:     "user-123",
				BookmarkID: 1,
				Score:      0.8,
				ReasonType: "collaborative",
			},
			wantErr: nil,
		},
		{
			name: "Invalid score too high",
			recommendation: BookmarkRecommendation{
				UserID:     "user-123",
				BookmarkID: 1,
				Score:      1.5,
				ReasonType: "collaborative",
			},
			wantErr: ErrInvalidScore,
		},
		{
			name: "Invalid reason type",
			recommendation: BookmarkRecommendation{
				UserID:     "user-123",
				BookmarkID: 1,
				Score:      0.8,
				ReasonType: "invalid",
			},
			wantErr: ErrInvalidReasonType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.recommendation.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
