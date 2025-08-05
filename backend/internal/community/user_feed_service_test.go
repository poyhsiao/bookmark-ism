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

// UserFeedServiceTestSuite tests the UserFeedService
type UserFeedServiceTestSuite struct {
	suite.Suite
	service   *UserFeedService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *UserFeedServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()

	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	suite.service = NewUserFeedService(suite.mockDB, suite.mockRedis, jsonHelper, logger)
}

func (suite *UserFeedServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// Test GenerateUserFeed - Success case
func (suite *UserFeedServiceTestSuite) TestGenerateUserFeed_Success() {
	request := &FeedRequest{
		UserID:     "user-123",
		Limit:      20,
		Offset:     0,
		SourceType: "all",
	}

	// Mock finding feed items
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "user-123"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Offset", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Return(&gorm.DB{Error: nil})

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), feed)
}

// Test GenerateUserFeed - Invalid user ID
func (suite *UserFeedServiceTestSuite) TestGenerateUserFeed_InvalidUserID() {
	request := &FeedRequest{
		UserID: "", // Invalid empty user ID
		Limit:  20,
	}

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), feed)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test GenerateUserFeed - Default limit
func (suite *UserFeedServiceTestSuite) TestGenerateUserFeed_DefaultLimit() {
	request := &FeedRequest{
		UserID: "user-123",
		Limit:  0, // Should default to 20
	}

	// Mock finding feed items with default limit
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "user-123"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", 20).Return(suite.mockDB)
	suite.mockDB.On("Offset", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Return(&gorm.DB{Error: nil})

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), feed)
}

// Test GenerateUserFeed - Specific source type
func (suite *UserFeedServiceTestSuite) TestGenerateUserFeed_SpecificSourceType() {
	request := &FeedRequest{
		UserID:     "user-123",
		Limit:      20,
		Offset:     0,
		SourceType: "trending",
	}

	// Mock finding feed items with specific source type
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "user-123"
	})).Return(suite.mockDB)
	suite.mockDB.On("Where", "source_type = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "trending"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Offset", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Return(&gorm.DB{Error: nil})

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), feed)
}

// Test GenerateUserFeed - Database error
func (suite *UserFeedServiceTestSuite) TestGenerateUserFeed_DatabaseError() {
	request := &FeedRequest{
		UserID: "user-123",
		Limit:  20,
	}

	// Mock database error
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "user-123"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Offset", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Return(&gorm.DB{Error: assert.AnError})

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), feed)
	assert.Contains(suite.T(), err.Error(), "failed to get user feed")
}

// Test GenerateUserFeed - Large limit (should be capped)
func (suite *UserFeedServiceTestSuite) TestGenerateUserFeed_LargeLimit() {
	request := &FeedRequest{
		UserID: "user-123",
		Limit:  150, // Should be capped to 100
	}

	// Mock finding feed items with capped limit
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "user-123"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", 20).Return(suite.mockDB) // Should use default 20, not 100
	suite.mockDB.On("Offset", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Return(&gorm.DB{Error: nil})

	feed, err := suite.service.GenerateUserFeed(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), feed)
}

// Run the test suite
func TestUserFeedServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserFeedServiceTestSuite))
}

// Simple unit test for feed response conversion
func TestUserFeedService_FeedResponseConversion(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	service := NewUserFeedService(mockDB, mockRedis, jsonHelper, logger)

	ctx := context.Background()
	request := &FeedRequest{
		UserID: "user-123",
		Limit:  5,
	}

	// Mock database query with specific feed items
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Order", mock.Anything).Return(mockDB)
	mockDB.On("Limit", mock.Anything).Return(mockDB)
	mockDB.On("Offset", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]community.UserFeed"), mock.Anything).Run(func(args mock.Arguments) {
		feedPtr := args.Get(0).(*[]UserFeed)
		*feedPtr = []UserFeed{
			{
				BookmarkID: 1,
				SourceType: "trending",
				SourceID:   "daily",
				Score:      0.9,
				Position:   1,
			},
			{
				BookmarkID: 2,
				SourceType: "following",
				SourceID:   "user-456",
				Score:      0.8,
				Position:   2,
			},
		}
	}).Return(&gorm.DB{Error: nil})

	feed, err := service.GenerateUserFeed(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, feed)
	assert.Len(t, feed, 2)

	// Verify first feed item
	assert.Equal(t, uint(1), feed[0].BookmarkID)
	assert.Equal(t, "trending", feed[0].SourceType)
	assert.Equal(t, "daily", feed[0].SourceID)
	assert.Equal(t, 0.9, feed[0].Score)
	assert.Equal(t, 1, feed[0].Position)

	// Verify second feed item
	assert.Equal(t, uint(2), feed[1].BookmarkID)
	assert.Equal(t, "following", feed[1].SourceType)
	assert.Equal(t, "user-456", feed[1].SourceID)
	assert.Equal(t, 0.8, feed[1].Score)
	assert.Equal(t, 2, feed[1].Position)

	mockDB.AssertExpectations(t)
}
