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

// RecommendationServiceTestSuite tests the RecommendationService
type RecommendationServiceTestSuite struct {
	suite.Suite
	service   *RecommendationService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *RecommendationServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()

	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	suite.service = NewRecommendationService(suite.mockDB, suite.mockRedis, jsonHelper, logger)
}

func (suite *RecommendationServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

// Test GetRecommendations - Success case
func (suite *RecommendationServiceTestSuite) TestGetRecommendations_Success() {
	request := &RecommendationRequest{
		UserID:    "user-123",
		Limit:     10,
		Algorithm: "collaborative",
		Context:   "homepage",
	}

	// Mock cache miss
	suite.mockRedis.On("Get", suite.ctx, "recommendations:user-123:collaborative:homepage").Return("", assert.AnError)

	// Mock finding existing recommendations
	suite.mockDB.On("Where", mock.MatchedBy(func(query interface{}) bool {
		return query == "user_id = ? AND expires_at > ?"
	}), mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Where", "reason_type = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "collaborative"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.BookmarkRecommendation"), mock.Anything).Run(func(args mock.Arguments) {
		// Populate with test data
		recommendations := args.Get(0).(*[]BookmarkRecommendation)
		*recommendations = []BookmarkRecommendation{
			{
				BookmarkID: 1,
				Score:      0.9,
				ReasonType: "collaborative",
				ReasonData: `{"similar_users": ["user1", "user2"]}`,
			},
		}
	}).Return(&gorm.DB{Error: nil})

	// Mock cache set
	suite.mockRedis.On("Set", suite.ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	recommendations, err := suite.service.GetRecommendations(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)
	assert.Len(suite.T(), recommendations, 1)
	assert.Equal(suite.T(), uint(1), recommendations[0].BookmarkID)
}

// Test GetRecommendations - Cache hit
func (suite *RecommendationServiceTestSuite) TestGetRecommendations_CacheHit() {
	request := &RecommendationRequest{
		UserID:    "user-123",
		Limit:     10,
		Algorithm: "collaborative",
		Context:   "homepage",
	}

	// Mock cache hit
	cachedData := `[{"bookmark_id":1,"score":0.9,"reason_type":"collaborative","reason_text":"Users with similar interests also liked this"}]`
	suite.mockRedis.On("Get", suite.ctx, "recommendations:user-123:collaborative:homepage").Return(cachedData, nil)

	recommendations, err := suite.service.GetRecommendations(suite.ctx, request)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)
	assert.Len(suite.T(), recommendations, 1)
	assert.Equal(suite.T(), uint(1), recommendations[0].BookmarkID)

	// Verify database was NOT called (cache hit)
	suite.mockDB.AssertNotCalled(suite.T(), "Where")
}

// Test GetRecommendations - Invalid user ID
func (suite *RecommendationServiceTestSuite) TestGetRecommendations_InvalidUserID() {
	request := &RecommendationRequest{
		UserID: "", // Invalid empty user ID
		Limit:  10,
	}

	recommendations, err := suite.service.GetRecommendations(suite.ctx, request)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), recommendations)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test GenerateRecommendations - Success case
func (suite *RecommendationServiceTestSuite) TestGenerateRecommendations_Success() {
	userID := "user-123"
	algorithm := "collaborative"

	// Mock finding user behaviors
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == userID
	})).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Run(func(args mock.Arguments) {
		behaviorsPtr := args.Get(0).(*[]UserBehavior)
		*behaviorsPtr = []UserBehavior{
			{UserID: userID, BookmarkID: 1, ActionType: "view"},
		}
	}).Return(&gorm.DB{Error: nil})

	// Mock creating recommendations
	suite.mockDB.On("Create", mock.AnythingOfType("*community.BookmarkRecommendation")).Return(&gorm.DB{Error: nil})

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	assert.NoError(suite.T(), err)
}

// Test GenerateRecommendations - Insufficient data
func (suite *RecommendationServiceTestSuite) TestGenerateRecommendations_InsufficientData() {
	userID := "user-123"
	algorithm := "collaborative"

	// Mock finding empty user behaviors (insufficient data)
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == userID
	})).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Run(func(args mock.Arguments) {
		behaviorsPtr := args.Get(0).(*[]UserBehavior)
		*behaviorsPtr = []UserBehavior{} // Empty slice
	}).Return(&gorm.DB{Error: nil})

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInsufficientData, err)

	// Verify no recommendations are created when there's insufficient data
	suite.mockDB.AssertNotCalled(suite.T(), "Create", mock.AnythingOfType("*community.BookmarkRecommendation"))
}

// Test GenerateRecommendations - Invalid algorithm
func (suite *RecommendationServiceTestSuite) TestGenerateRecommendations_InvalidAlgorithm() {
	userID := "user-123"
	algorithm := "invalid"

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidAlgorithm, err)
}

// Test GenerateRecommendations - Invalid user ID
func (suite *RecommendationServiceTestSuite) TestGenerateRecommendations_InvalidUserID() {
	userID := ""
	algorithm := "collaborative"

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidUserID, err)
}

// Test different recommendation algorithms - simplified to test basic algorithms
func (suite *RecommendationServiceTestSuite) TestGenerateRecommendations_DifferentAlgorithms() {
	userID := "user-123"
	algorithms := []string{"collaborative", "content_based"}

	for _, algorithm := range algorithms {
		suite.SetupTest() // Reset mocks for each test

		// Mock finding user behaviors
		suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
			return len(args) == 1 && args[0] == userID
		})).Return(suite.mockDB)
		suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Run(func(args mock.Arguments) {
			behaviorsPtr := args.Get(0).(*[]UserBehavior)
			*behaviorsPtr = []UserBehavior{
				{UserID: userID, BookmarkID: 1, ActionType: "view"},
			}
		}).Return(&gorm.DB{Error: nil})

		// Mock creating recommendations
		suite.mockDB.On("Create", mock.AnythingOfType("*community.BookmarkRecommendation")).Return(&gorm.DB{Error: nil}).Maybe()

		err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

		assert.NoError(suite.T(), err, "Failed for algorithm: %s", algorithm)

		suite.TearDownTest() // Verify expectations for each test
	}
}

// Test trending algorithm specifically
func (suite *RecommendationServiceTestSuite) TestGenerateRecommendations_TrendingAlgorithm() {
	userID := "user-123"
	algorithm := "trending"

	// Mock finding user behaviors
	suite.mockDB.On("Where", "user_id = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == userID
	})).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.UserBehavior"), mock.Anything).Run(func(args mock.Arguments) {
		behaviorsPtr := args.Get(0).(*[]UserBehavior)
		*behaviorsPtr = []UserBehavior{
			{UserID: userID, BookmarkID: 1, ActionType: "view"},
		}
	}).Return(&gorm.DB{Error: nil})

	// Mock trending bookmarks query
	suite.mockDB.On("Where", "time_window = ?", mock.MatchedBy(func(args []interface{}) bool {
		return len(args) == 1 && args[0] == "daily"
	})).Return(suite.mockDB)
	suite.mockDB.On("Order", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Limit", mock.Anything).Return(suite.mockDB)
	suite.mockDB.On("Find", mock.AnythingOfType("*[]community.TrendingBookmark"), mock.Anything).Return(&gorm.DB{Error: nil})

	// Mock creating recommendations
	suite.mockDB.On("Create", mock.AnythingOfType("*community.BookmarkRecommendation")).Return(&gorm.DB{Error: nil}).Maybe()

	err := suite.service.GenerateRecommendations(suite.ctx, userID, algorithm)

	assert.NoError(suite.T(), err)
}

// Run the test suite
func TestRecommendationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RecommendationServiceTestSuite))
}

// Simple unit test for reason text generation
func TestRecommendationService_GenerateReasonText(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	logger := zap.NewNop()
	jsonHelper := NewJSONHelper()
	service := NewRecommendationService(mockDB, mockRedis, jsonHelper, logger)

	tests := []struct {
		reasonType string
		expected   string
	}{
		{"collaborative", "Users with similar interests also liked this"},
		{"content_based", "Similar to bookmarks you've saved"},
		{"trending", "Trending in your network"},
		{"popularity", "Popular among all users"},
		{"hybrid", "Recommended based on your activity and trends"},
		{"unknown", "Recommended for you"},
	}

	for _, tt := range tests {
		result := service.generateReasonText(tt.reasonType)
		assert.Equal(t, tt.expected, result, "Failed for reason type: %s", tt.reasonType)
	}
}
