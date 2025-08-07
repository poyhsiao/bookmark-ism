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

// MockWorkerPool for testing
type MockWorkerPool struct {
	mock.Mock
}

func (m *MockWorkerPool) Submit(job worker.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

// BehaviorTrackingServiceTestSuite tests the BehaviorTrackingService
type BehaviorTrackingServiceTestSuite struct {
	suite.Suite
	service        *BehaviorTrackingService
	mockDB         *MockDB
	mockRedis      *MockRedisClient
	mockWorkerPool *MockWorkerPool
	ctx            context.Context
}

func (suite *BehaviorTrackingServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.mockWorkerPool = new(MockWorkerPool)
	suite.ctx = context.Background()
	logger := zaptest.NewLogger(suite.T())

	// Create real worker pool for testing
	workerPool := worker.NewWorkerPool(1, 5, logger)

	// Create dependent services
	jsonHelper := NewJSONHelper()
	socialMetrics := NewSocialMetricsService(suite.mockDB, suite.mockRedis, jsonHelper, logger)
	trending := NewTrendingService(suite.mockDB, suite.mockRedis, jsonHelper, logger)

	suite.service = NewBehaviorTrackingService(
		suite.mockDB,
		suite.mockRedis,
		workerPool,
		socialMetrics,
		trending,
		logger,
	)
}

func (suite *BehaviorTrackingServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

func (suite *BehaviorTrackingServiceTestSuite) TestTrackUserBehavior_Success() {
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

func (suite *BehaviorTrackingServiceTestSuite) TestTrackUserBehavior_InvalidUserID() {
	request := &BehaviorTrackingRequest{
		UserID:     "", // Invalid empty user ID
		BookmarkID: 1,
		ActionType: "view",
	}

	err := suite.service.TrackUserBehavior(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

func (suite *BehaviorTrackingServiceTestSuite) TestTrackUserBehavior_InvalidBookmarkID() {
	request := &BehaviorTrackingRequest{
		UserID:     "user-123",
		BookmarkID: 0, // Invalid bookmark ID
		ActionType: "view",
	}

	err := suite.service.TrackUserBehavior(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidBookmarkID, err)
}

func (suite *BehaviorTrackingServiceTestSuite) TestTrackUserBehavior_InvalidActionType() {
	request := &BehaviorTrackingRequest{
		UserID:     "user-123",
		BookmarkID: 1,
		ActionType: "invalid", // Invalid action type
	}

	err := suite.service.TrackUserBehavior(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidActionType, err)
}

func TestBehaviorTrackingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BehaviorTrackingServiceTestSuite))
}

// Unit tests for helper methods
func TestBehaviorTrackingService_ValidateRequest(t *testing.T) {
	service := &BehaviorTrackingService{}

	tests := []struct {
		name     string
		request  *BehaviorTrackingRequest
		expected error
	}{
		{
			name: "Valid request",
			request: &BehaviorTrackingRequest{
				UserID:     "user-123",
				BookmarkID: 1,
				ActionType: "view",
			},
			expected: nil,
		},
		{
			name: "Invalid user ID",
			request: &BehaviorTrackingRequest{
				UserID:     "",
				BookmarkID: 1,
				ActionType: "view",
			},
			expected: ErrInvalidUserID,
		},
		{
			name: "Invalid bookmark ID",
			request: &BehaviorTrackingRequest{
				UserID:     "user-123",
				BookmarkID: 0,
				ActionType: "view",
			},
			expected: ErrInvalidBookmarkID,
		},
		{
			name: "Invalid action type",
			request: &BehaviorTrackingRequest{
				UserID:     "user-123",
				BookmarkID: 1,
				ActionType: "invalid",
			},
			expected: ErrInvalidActionType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRequest(tt.request)
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestBehaviorTrackingService_CreateBehaviorRecord(t *testing.T) {
	service := &BehaviorTrackingService{}

	request := &BehaviorTrackingRequest{
		UserID:     "user-123",
		BookmarkID: 1,
		ActionType: "view",
		Duration:   30,
		Context:    "homepage",
		Metadata:   map[string]interface{}{"source": "recommendation"},
	}

	behavior, err := service.createBehaviorRecord(request)

	assert.NoError(t, err)
	assert.NotNil(t, behavior)
	assert.Equal(t, request.UserID, behavior.UserID)
	assert.Equal(t, request.BookmarkID, behavior.BookmarkID)
	assert.Equal(t, request.ActionType, behavior.ActionType)
	assert.Equal(t, request.Duration, behavior.Duration)
	assert.Equal(t, request.Context, behavior.Context)
	assert.NotEmpty(t, behavior.Metadata) // Should contain serialized JSON
}

func TestBehaviorTrackingService_IsSignificantAction(t *testing.T) {
	service := &BehaviorTrackingService{}

	tests := []struct {
		actionType string
		expected   bool
	}{
		{"view", true},
		{"click", true},
		{"save", true},
		{"share", true},
		{"like", true},
		{"dismiss", false},
		{"report", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.actionType, func(t *testing.T) {
			result := service.isSignificantAction(tt.actionType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
