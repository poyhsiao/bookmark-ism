package community

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

// RefactoredIntegrationTestSuite tests the integration between refactored services
type RefactoredIntegrationTestSuite struct {
	suite.Suite
	service   *RefactoredService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *RefactoredIntegrationTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()
	logger := zaptest.NewLogger(suite.T())

	// Create worker pool for testing
	workerPool := worker.NewWorkerPool(2, 10, logger)
	workerPool.Start()

	suite.service = NewRefactoredService(
		suite.mockDB,
		suite.mockRedis,
		workerPool,
		logger,
	)
}

func (suite *RefactoredIntegrationTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// TDD Test: Write failing test first, then implement
func (suite *RefactoredIntegrationTestSuite) TestCompleteUserJourney_TrackBehaviorAndGetRecommendations() {
	// This test demonstrates a complete user journey using the refactored services
	userID := "user-123"
	bookmarkID := uint(1)

	// Step 1: Track user behavior (should trigger async social metrics update)
	behaviorRequest := &BehaviorTrackingRequest{
		UserID:     userID,
		BookmarkID: bookmarkID,
		ActionType: "view",
		Duration:   30,
		Context:    "homepage",
	}

	// Mock behavior tracking
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserBehavior")).Return(&gorm.DB{Error: nil})

	err := suite.service.TrackUserBehavior(suite.ctx, behaviorRequest)
	assert.NoError(suite.T(), err)

	// Step 2: Get recommendations (should use cache and database)
	recommendationRequest := &RecommendationRequest{
		UserID:    userID,
		Limit:     10,
		Algorithm: "collaborative",
		Context:   "homepage",
	}

	// Mock cache miss and database query for recommendations
	suite.mockRedis.On("Get", suite.ctx, "recommendations:user-123:collaborative:homepage").Return("", nil)
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.BookmarkRecommendation"), mock.Anything).Return(&gorm.DB{Error: nil})
	suite.mockRedis.On("Set", suite.ctx, "recommendations:user-123:collaborative:homepage", mock.Anything, 15*time.Minute).Return(nil)

	recommendations, err := suite.service.GetRecommendations(suite.ctx, recommendationRequest)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)

	// Step 3: Get social metrics for the bookmark
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: nil})

	metrics, err := suite.service.GetSocialMetrics(suite.ctx, bookmarkID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), metrics)
	assert.Equal(suite.T(), bookmarkID, metrics.BookmarkID)
}

// TDD Test: Test error handling across services
func (suite *RefactoredIntegrationTestSuite) TestErrorPropagation_AcrossServices() {
	// Test that errors are properly propagated from domain services to the main service

	// Test invalid user ID propagation
	err := suite.service.TrackUserBehavior(suite.ctx, &BehaviorTrackingRequest{
		UserID:     "", // Invalid
		BookmarkID: 1,
		ActionType: "view",
	})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidUserID, err)

	// Test invalid bookmark ID propagation
	metrics, err := suite.service.GetSocialMetrics(suite.ctx, 0) // Invalid bookmark ID
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), metrics)
	assert.Equal(suite.T(), ErrInvalidBookmarkID, err)
}

// TDD Test: Test caching behavior across services
func (suite *RefactoredIntegrationTestSuite) TestCaching_ConsistencyAcrossServices() {
	userID := "user-123"

	// Test user stats caching
	// First call should hit database and cache the result
	suite.mockRedis.On("Get", suite.ctx, "user_stats:user-123").Return("", nil) // Cache miss
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: nil})
	suite.mockRedis.On("Set", suite.ctx, "user_stats:user-123", mock.Anything, 30*time.Minute).Return(nil)

	stats1, err := suite.service.GetUserStats(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats1)

	// Follow a user (should clear cache)
	followRequest := &FollowRequest{FollowingID: "user-456"}
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserFollow")).Return(&gorm.DB{Error: nil})
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-123"}).Return(nil)
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-456"}).Return(nil)

	err = suite.service.FollowUser(suite.ctx, userID, followRequest)
	assert.NoError(suite.T(), err)
}

// TDD Test: Test service composition and delegation
func (suite *RefactoredIntegrationTestSuite) TestServiceComposition_ProperDelegation() {
	// This test ensures that the RefactoredService properly delegates to the correct domain services
	// and that each domain service handles its specific responsibilities

	// Test that behavior tracking delegates to BehaviorTrackingService
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserBehavior")).Return(&gorm.DB{Error: nil})
	err := suite.service.TrackUserBehavior(suite.ctx, &BehaviorTrackingRequest{
		UserID: "user-123", BookmarkID: 1, ActionType: "view",
	})
	assert.NoError(suite.T(), err)

	// Test that social metrics delegates to SocialMetricsService
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: nil})
	_, err = suite.service.GetSocialMetrics(suite.ctx, 1)
	assert.NoError(suite.T(), err)

	// Test that trending delegates to TrendingService
	suite.mockDB.On("Where", mock.Anything, mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.TrendingBookmark"), mock.Anything).Return(&gorm.DB{Error: nil})
	_, err = suite.service.GetTrendingBookmarksInternal(suite.ctx, &TrendingRequest{TimeWindow: "daily", Limit: 10})
	assert.NoError(suite.T(), err)
}

func TestRefactoredIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RefactoredIntegrationTestSuite))
}

// Benchmark test to compare performance of refactored vs original service
func BenchmarkRefactoredService_TrackUserBehavior(b *testing.B) {
	mockDB := new(MockDB)
	mockRedis := new(MockRedisClient)
	logger := zaptest.NewLogger(b)
	workerPool := worker.NewWorkerPool(2, 10, logger)

	service := NewRefactoredService(mockDB, mockRedis, workerPool, logger)
	ctx := context.Background()

	request := &BehaviorTrackingRequest{
		UserID:     "user-123",
		BookmarkID: 1,
		ActionType: "view",
		Duration:   30,
		Context:    "homepage",
	}

	// Mock the database call
	mockDB.On("Create", mock.AnythingOfType("*community.UserBehavior")).Return(&gorm.DB{Error: nil})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.TrackUserBehavior(ctx, request)
	}
}

// Test helper functions and utilities
func TestRefactoredService_HelperIntegration(t *testing.T) {
	// Test that helpers are properly integrated and used by services
	jsonHelper := NewJSONHelper()
	// cacheHelper := NewCacheHelper(new(MockRedisClient), jsonHelper)
	configHelper := NewConfigHelper()
	validationHelper := NewValidationHelper()

	// Test JSON helper
	data := map[string]string{"test": "value"}
	jsonStr, err := jsonHelper.MarshalToString(data)
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "test")

	// Test config helper
	assert.True(t, configHelper.ValidateTimeWindow("daily"))
	assert.False(t, configHelper.ValidateTimeWindow("invalid"))

	// Test validation helper
	assert.NoError(t, validationHelper.ValidateUserID("user-123"))
	assert.Equal(t, ErrInvalidUserID, validationHelper.ValidateUserID(""))

	// Test that limit validation works correctly
	assert.Equal(t, 10, validationHelper.ValidateLimit(10))
	assert.Equal(t, 20, validationHelper.ValidateLimit(0))   // Default
	assert.Equal(t, 20, validationHelper.ValidateLimit(150)) // Over max
}
