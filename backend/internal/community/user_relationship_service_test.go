package community

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRelationshipServiceTestSuite tests the UserRelationshipService
type UserRelationshipServiceTestSuite struct {
	suite.Suite
	service   *UserRelationshipService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *UserRelationshipServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()

	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(suite.mockRedis, jsonHelper)
	suite.service = NewUserRelationshipService(suite.mockDB, suite.mockRedis, cacheHelper, logger)
}

func (suite *UserRelationshipServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// Test FollowUser - Success case
func (suite *UserRelationshipServiceTestSuite) TestFollowUser_Success() {
	request := &FollowRequest{
		FollowingID: "user-456",
	}
	followerID := "user-123"

	// Mock check for existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	// Mock successful creation
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserFollow")).Return(&gorm.DB{Error: nil})

	// Mock cache clearing
	suite.mockRedis.On("Del", suite.ctx, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == "user_stats:user-123"
	})).Return(nil)
	suite.mockRedis.On("Del", suite.ctx, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == "user_stats:user-456"
	})).Return(nil)

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.NoError(suite.T(), err)
}

// Test FollowUser - Already following
func (suite *UserRelationshipServiceTestSuite) TestFollowUser_AlreadyFollowing() {
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

// Test FollowUser - Cannot follow self
func (suite *UserRelationshipServiceTestSuite) TestFollowUser_CannotFollowSelf() {
	request := &FollowRequest{
		FollowingID: "user-123",
	}
	followerID := "user-123"

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrCannotFollowSelf, err)
}

// Test FollowUser - Invalid follower ID
func (suite *UserRelationshipServiceTestSuite) TestFollowUser_InvalidFollowerID() {
	request := &FollowRequest{
		FollowingID: "user-456",
	}
	followerID := ""

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidFollowerID, err)
}

// Test FollowUser - Invalid following ID
func (suite *UserRelationshipServiceTestSuite) TestFollowUser_InvalidFollowingID() {
	request := &FollowRequest{
		FollowingID: "",
	}
	followerID := "user-123"

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidFollowingID, err)
}

// Test UnfollowUser - Success case
func (suite *UserRelationshipServiceTestSuite) TestUnfollowUser_Success() {
	followingID := "user-456"
	followerID := "user-123"

	// Mock existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock successful deletion
	suite.mockDB.On("Delete", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock cache clearing
	suite.mockRedis.On("Del", suite.ctx, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == "user_stats:user-123"
	})).Return(nil)
	suite.mockRedis.On("Del", suite.ctx, mock.MatchedBy(func(keys []string) bool {
		return len(keys) == 1 && keys[0] == "user_stats:user-456"
	})).Return(nil)

	err := suite.service.UnfollowUser(suite.ctx, followerID, followingID)

	assert.NoError(suite.T(), err)
}

// Test UnfollowUser - Not following
func (suite *UserRelationshipServiceTestSuite) TestUnfollowUser_NotFollowing() {
	followingID := "user-456"
	followerID := "user-123"

	// Mock no existing follow relationship
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	err := suite.service.UnfollowUser(suite.ctx, followerID, followingID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrNotFollowing, err)
}

// Test UnfollowUser - Invalid follower ID
func (suite *UserRelationshipServiceTestSuite) TestUnfollowUser_InvalidFollowerID() {
	followingID := "user-456"
	followerID := ""

	err := suite.service.UnfollowUser(suite.ctx, followerID, followingID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidFollowerID, err)
}

// Test UnfollowUser - Invalid following ID
func (suite *UserRelationshipServiceTestSuite) TestUnfollowUser_InvalidFollowingID() {
	followingID := ""
	followerID := "user-123"

	err := suite.service.UnfollowUser(suite.ctx, followerID, followingID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidFollowingID, err)
}

// Test GetUserStats - Success case
func (suite *UserRelationshipServiceTestSuite) TestGetUserStats_Success() {
	userID := "user-123"

	// Mock cache miss
	suite.mockRedis.On("Get", suite.ctx, "user_stats:user-123").Return("", assert.AnError)

	// Mock followers query
	suite.mockDB.On("Where", "following_id = ? AND status = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 2 && args[0] == userID && args[1] == "active"
	})).Return(suite.mockDB).Once()
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Run(func(args mock.Arguments) {
		followersPtr := args.Get(0).(*[]UserFollow)
		*followersPtr = make([]UserFollow, 5) // 5 followers
	}).Return(&gorm.DB{Error: nil}).Once()

	// Mock following query
	suite.mockDB.On("Where", "follower_id = ? AND status = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 2 && args[0] == userID && args[1] == "active"
	})).Return(suite.mockDB).Once()
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Run(func(args mock.Arguments) {
		followingPtr := args.Get(0).(*[]UserFollow)
		*followingPtr = make([]UserFollow, 3) // 3 following
	}).Return(&gorm.DB{Error: nil}).Once()

	// Mock cache set
	suite.mockRedis.On("Set", suite.ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	stats, err := suite.service.GetUserStats(suite.ctx, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), userID, stats.UserID)
	assert.Equal(suite.T(), 5, stats.FollowersCount)
	assert.Equal(suite.T(), 3, stats.FollowingCount)
}

// Test GetUserStats - Cache hit
func (suite *UserRelationshipServiceTestSuite) TestGetUserStats_CacheHit() {
	userID := "user-123"

	// Mock cache hit
	cachedData := `{"user_id":"user-123","followers_count":10,"following_count":5,"influence_score":7.5}`
	suite.mockRedis.On("Get", suite.ctx, "user_stats:user-123").Return(cachedData, nil)

	stats, err := suite.service.GetUserStats(suite.ctx, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), userID, stats.UserID)
	assert.Equal(suite.T(), 10, stats.FollowersCount)
	assert.Equal(suite.T(), 5, stats.FollowingCount)

	// Verify database was NOT called (cache hit)
	suite.mockDB.AssertNotCalled(suite.T(), "Where")
}

// Test GetUserStats - Invalid user ID
func (suite *UserRelationshipServiceTestSuite) TestGetUserStats_InvalidUserID() {
	userID := ""

	stats, err := suite.service.GetUserStats(suite.ctx, userID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), stats)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Run the test suite
func TestUserRelationshipServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserRelationshipServiceTestSuite))
}

// Simple unit test for influence score calculation
func TestUserRelationshipService_InfluenceScoreCalculation(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(mockRedis, jsonHelper)
	service := NewUserRelationshipService(mockDB, mockRedis, cacheHelper, logger)

	ctx := context.Background()
	userID := "user-123"

	// Mock cache miss
	mockRedis.On("Get", ctx, "user_stats:user-123").Return("", assert.AnError)

	// Mock database queries with specific data
	mockDB.On("Where", "following_id = ? AND status = ?", []interface{}{userID, "active"}).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Run(func(args mock.Arguments) {
		followersPtr := args.Get(0).(*[]UserFollow)
		*followersPtr = make([]UserFollow, 10) // 10 followers
	}).Return(&gorm.DB{Error: nil}).Once()

	mockDB.On("Where", "follower_id = ? AND status = ?", []interface{}{userID, "active"}).Return(mockDB).Once()
	mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Run(func(args mock.Arguments) {
		followingPtr := args.Get(0).(*[]UserFollow)
		*followingPtr = make([]UserFollow, 5) // 5 following
	}).Return(&gorm.DB{Error: nil}).Once()

	// Mock cache set
	mockRedis.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	stats, err := service.GetUserStats(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 10, stats.FollowersCount)
	assert.Equal(t, 5, stats.FollowingCount)
	// Influence score = followers * 0.7 + engagement * 0.3 = 10 * 0.7 + 0 * 0.3 = 7.0
	assert.Equal(t, 7.0, stats.InfluenceScore)

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}
