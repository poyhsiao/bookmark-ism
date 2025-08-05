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

// TrendingServiceTestSuite tests the TrendingService
type TrendingServiceTestSuite struct {
	suite.Suite
	service   *TrendingService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *TrendingServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()

	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	suite.service = NewTrendingService(suite.mockDB, suite.mockRedis, jsonHelper, logger)
}

func (suite *TrendingServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// Test GetTrendingBookmarksInternal - Success case
func (suite *TrendingServiceTestSuite) TestGetTrendingBookmarksInternal_Success() {
	request := &TrendingRequest{
		TimeWindow: "daily",
		Limit:      20,
		MinScore:   0.5,
	}

	// Mock finding trending bookmarks
	suite.mockDB.On("Where", "time_window = ?", "daily").Return(suite.mockDB)
	suite.mockDB.On("Where", "trending_score >= ?", 0.5).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.TrendingBookmark"), mock.Anything).Return(&gorm.DB{Error: nil})

	trending, err := suite.service.GetTrendingBookmarksInternal(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), trending)
}

// Test GetTrendingBookmarksInternal - Invalid time window
func (suite *TrendingServiceTestSuite) TestGetTrendingBookmarksInternal_InvalidTimeWindow() {
	request := &TrendingRequest{
		TimeWindow: "invalid",
		Limit:      20,
	}

	trending, err := suite.service.GetTrendingBookmarksInternal(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), trending)
	assert.Equal(suite.T(), ErrInvalidTimeWindow, err)
}

// Test CalculateTrendingScores - Success case
func (suite *TrendingServiceTestSuite) TestCalculateTrendingScores_Success() {
	timeWindow := "daily"

	// Mock finding user behaviors
	suite.mockDB.On("Where", "created_at >= ?", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Return(&gorm.DB{Error: nil})

	err := suite.service.CalculateTrendingScores(suite.ctx, timeWindow)

	assert.NoError(suite.T(), err)
}

// Test CalculateTrendingScores - Invalid time window
func (suite *TrendingServiceTestSuite) TestCalculateTrendingScores_InvalidTimeWindow() {
	timeWindow := "invalid"

	err := suite.service.CalculateTrendingScores(suite.ctx, timeWindow)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidTimeWindow, err)
}

// Test UpdateTrendingCache - Success case
func (suite *TrendingServiceTestSuite) TestUpdateTrendingCache_Success() {
	bookmarkID := uint(1)
	actionType := "view"

	// Mock Redis ZAdd operation
	suite.mockRedis.On("ZAdd", suite.ctx, "trending:view", mock.Anything, mock.Anything).Return(nil)

	err := suite.service.UpdateTrendingCache(suite.ctx, bookmarkID, actionType)

	assert.NoError(suite.T(), err)
}

// Run the test suite
func TestTrendingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TrendingServiceTestSuite))
}
