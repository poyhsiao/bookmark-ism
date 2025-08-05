package community

import (
	"context"
	"testing"

	"bookmark-sync-service/backend/pkg/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

// RefactoredServiceTestSuite tests the RefactoredService
type RefactoredServiceTestSuite struct {
	suite.Suite
	service   *RefactoredService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *RefactoredServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()
	logger := zaptest.NewLogger(suite.T())

	// Create worker pool for testing
	workerPool := worker.NewWorkerPool(1, 5, logger)

	suite.service = NewRefactoredService(
		suite.mockDB,
		suite.mockRedis,
		workerPool,
		logger,
	)
}

func (suite *RefactoredServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// Test that RefactoredService properly delegates to domain services
func (suite *RefactoredServiceTestSuite) TestTrackUserBehavior_DelegatesToBehaviorTrackingService() {
	request := &BehaviorTrackingRequest{
		UserID:     "user-123",
		BookmarkID: 1,
		ActionType: "view",
		Duration:   30,
		Context:    "homepage",
	}

	// Mock the database call that BehaviorTrackingService will make
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserBehavior")).Return(&gorm.DB{Error: nil})

	err := suite.service.TrackUserBehavior(suite.ctx, request)

	assert.NoError(suite.T(), err)
}

func (suite *RefactoredServiceTestSuite) TestFollowUser_DelegatesToUserRelationshipService() {
	request := &FollowRequest{
		FollowingID: "user-456",
	}
	followerID := "user-123"

	// Mock the database calls that UserRelationshipService will make
	suite.mockDB.On("First", mock.AnythingOfType("*community.UserFollow"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
	suite.mockDB.On("Create", mock.AnythingOfType("*community.UserFollow")).Return(&gorm.DB{Error: nil})
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-123"}).Return(nil)
	suite.mockRedis.On("Del", suite.ctx, []string{"user_stats:user-456"}).Return(nil)

	err := suite.service.FollowUser(suite.ctx, followerID, request)

	assert.NoError(suite.T(), err)
}

// Run the test suite
func TestRefactoredServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RefactoredServiceTestSuite))
}
