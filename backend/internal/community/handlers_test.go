package community

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock service for testing handlers
type MockService struct {
	mock.Mock
}

func (m *MockService) TrackUserBehavior(ctx context.Context, req *BehaviorTrackingRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockService) FollowUser(ctx context.Context, followerID string, req *FollowRequest) error {
	args := m.Called(ctx, followerID, req)
	return args.Error(0)
}

func (m *MockService) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	args := m.Called(ctx, followerID, followingID)
	return args.Error(0)
}

func (m *MockService) GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]RecommendationResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]RecommendationResponse), args.Error(1)
}

func (m *MockService) GetTrendingBookmarksInternal(ctx context.Context, req *TrendingRequest) ([]TrendingResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]TrendingResponse), args.Error(1)
}

func (m *MockService) GenerateUserFeed(ctx context.Context, req *FeedRequest) ([]UserFeedResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]UserFeedResponse), args.Error(1)
}

func (m *MockService) GetSocialMetrics(ctx context.Context, bookmarkID uint) (*SocialMetricsResponse, error) {
	args := m.Called(ctx, bookmarkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SocialMetricsResponse), args.Error(1)
}

func (m *MockService) GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserStatsResponse), args.Error(1)
}

func (m *MockService) GenerateRecommendations(ctx context.Context, userID, algorithm string) error {
	args := m.Called(ctx, userID, algorithm)
	return args.Error(0)
}

func (m *MockService) CalculateTrendingScores(ctx context.Context, timeWindow string) error {
	args := m.Called(ctx, timeWindow)
	return args.Error(0)
}

func (m *MockService) UpdateSocialMetrics(ctx context.Context, bookmarkID uint, actionType string) error {
	args := m.Called(ctx, bookmarkID, actionType)
	return args.Error(0)
}

// Test Suite
type CommunityHandlerTestSuite struct {
	suite.Suite
	handler     *Handler
	mockService *MockService
	router      *gin.Engine
}

func (suite *CommunityHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockService = new(MockService)
	suite.handler = NewHandler(suite.mockService)

	suite.router = gin.New()

	// Add auth middleware mock
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})

	api := suite.router.Group("/api/v1")
	suite.handler.RegisterRoutes(api)
}

func (suite *CommunityHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// Test TrackBehavior endpoint
func (suite *CommunityHandlerTestSuite) TestTrackBehavior() {
	request := BehaviorTrackingRequest{
		BookmarkID: 1,
		ActionType: "view",
		Duration:   30,
		Context:    "homepage",
	}

	suite.mockService.On("TrackUserBehavior", mock.Anything, mock.MatchedBy(func(req *BehaviorTrackingRequest) bool {
		return req.UserID == "test-user-123" && req.BookmarkID == 1 && req.ActionType == "view"
	})).Return(nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/community/behavior", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Behavior tracked successfully", response["message"])
}

func (suite *CommunityHandlerTestSuite) TestTrackBehaviorInvalidRequest() {
	request := map[string]interface{}{
		"bookmark_id": "invalid", // Invalid type
		"action_type": "view",
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/community/behavior", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *CommunityHandlerTestSuite) TestTrackBehaviorServiceError() {
	request := BehaviorTrackingRequest{
		BookmarkID: 1,
		ActionType: "view",
	}

	suite.mockService.On("TrackUserBehavior", mock.Anything, mock.Anything).Return(ErrInvalidBookmarkID)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/community/behavior", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Test FollowUser endpoint
func (suite *CommunityHandlerTestSuite) TestFollowUser() {
	request := FollowRequest{
		FollowingID: "user-456",
	}

	suite.mockService.On("FollowUser", mock.Anything, "test-user-123", mock.MatchedBy(func(req *FollowRequest) bool {
		return req.FollowingID == "user-456"
	})).Return(nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/community/follow", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "User followed successfully", response["message"])
}

func (suite *CommunityHandlerTestSuite) TestFollowUserAlreadyFollowing() {
	request := FollowRequest{
		FollowingID: "user-456",
	}

	suite.mockService.On("FollowUser", mock.Anything, "test-user-123", mock.Anything).Return(ErrAlreadyFollowing)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/community/follow", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

// Test UnfollowUser endpoint
func (suite *CommunityHandlerTestSuite) TestUnfollowUser() {
	suite.mockService.On("UnfollowUser", mock.Anything, "test-user-123", "user-456").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/community/follow/user-456", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "User unfollowed successfully", response["message"])
}

func (suite *CommunityHandlerTestSuite) TestUnfollowUserNotFollowing() {
	suite.mockService.On("UnfollowUser", mock.Anything, "test-user-123", "user-456").Return(ErrNotFollowing)

	req := httptest.NewRequest("DELETE", "/api/v1/community/follow/user-456", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test GetRecommendations endpoint
func (suite *CommunityHandlerTestSuite) TestGetRecommendations() {
	expectedRecommendations := []RecommendationResponse{
		{
			BookmarkID: 1,
			Score:      0.8,
			ReasonType: "collaborative",
			ReasonText: "Users with similar interests also liked this",
		},
		{
			BookmarkID: 2,
			Score:      0.7,
			ReasonType: "trending",
			ReasonText: "Trending in your network",
		},
	}

	suite.mockService.On("GetRecommendations", mock.Anything, mock.MatchedBy(func(req *RecommendationRequest) bool {
		return req.UserID == "test-user-123" && req.Limit == 20
	})).Return(expectedRecommendations, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/recommendations", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), float64(2), response["total"])

	recommendations := response["recommendations"].([]interface{})
	assert.Len(suite.T(), recommendations, 2)
}

func (suite *CommunityHandlerTestSuite) TestGetRecommendationsWithParams() {
	expectedRecommendations := []RecommendationResponse{
		{BookmarkID: 1, Score: 0.8, ReasonType: "collaborative"},
	}

	suite.mockService.On("GetRecommendations", mock.Anything, mock.MatchedBy(func(req *RecommendationRequest) bool {
		return req.UserID == "test-user-123" && req.Limit == 10 && req.Algorithm == "collaborative"
	})).Return(expectedRecommendations, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/recommendations?limit=10&algorithm=collaborative", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Test GetTrending endpoint
func (suite *CommunityHandlerTestSuite) TestGetTrending() {
	expectedTrending := []TrendingResponse{
		{
			BookmarkID:     1,
			TrendingScore:  95.5,
			ViewCount:      1000,
			EngagementRate: 0.15,
			TimeWindow:     "daily",
		},
	}

	suite.mockService.On("GetTrendingBookmarksInternal", mock.Anything, mock.MatchedBy(func(req *TrendingRequest) bool {
		return req.TimeWindow == "daily" && req.Limit == 20
	})).Return(expectedTrending, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/trending", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), float64(1), response["total"])
	assert.Equal(suite.T(), "daily", response["time_window"])
}

func (suite *CommunityHandlerTestSuite) TestGetTrendingWithParams() {
	expectedTrending := []TrendingResponse{}

	suite.mockService.On("GetTrendingBookmarksInternal", mock.Anything, mock.MatchedBy(func(req *TrendingRequest) bool {
		return req.TimeWindow == "weekly" && req.Limit == 50 && req.MinScore == 10.0
	})).Return(expectedTrending, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/trending?time_window=weekly&limit=50&min_score=10.0", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Test GetFeed endpoint
func (suite *CommunityHandlerTestSuite) TestGetFeed() {
	expectedFeed := []UserFeedResponse{
		{
			BookmarkID: 1,
			SourceType: "following",
			SourceID:   "user-456",
			Score:      0.9,
			Position:   1,
		},
	}

	suite.mockService.On("GenerateUserFeed", mock.Anything, mock.MatchedBy(func(req *FeedRequest) bool {
		return req.UserID == "test-user-123" && req.Limit == 20 && req.SourceType == "all"
	})).Return(expectedFeed, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/feed", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), float64(1), response["total"])
	assert.Equal(suite.T(), "all", response["source_type"])
}

// Test GetSocialMetrics endpoint
func (suite *CommunityHandlerTestSuite) TestGetSocialMetrics() {
	expectedMetrics := &SocialMetricsResponse{
		BookmarkID:     1,
		TotalViews:     1000,
		TotalClicks:    150,
		TotalSaves:     50,
		TotalShares:    25,
		TotalLikes:     75,
		UniqueViewers:  800,
		EngagementRate: 0.15,
		ViralityScore:  125.0,
		QualityScore:   0.8,
	}

	suite.mockService.On("GetSocialMetrics", mock.Anything, uint(1)).Return(expectedMetrics, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/metrics/1", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response SocialMetricsResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), uint(1), response.BookmarkID)
	assert.Equal(suite.T(), 1000, response.TotalViews)
}

func (suite *CommunityHandlerTestSuite) TestGetSocialMetricsInvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/community/metrics/invalid", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *CommunityHandlerTestSuite) TestGetSocialMetricsNotFound() {
	suite.mockService.On("GetSocialMetrics", mock.Anything, uint(999)).Return(nil, ErrBookmarkNotFound)

	req := httptest.NewRequest("GET", "/api/v1/community/metrics/999", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test GetUserStats endpoint
func (suite *CommunityHandlerTestSuite) TestGetUserStats() {
	expectedStats := &UserStatsResponse{
		UserID:          "user-456",
		FollowersCount:  100,
		FollowingCount:  50,
		BookmarksCount:  200,
		PublicBookmarks: 150,
		TotalViews:      5000,
		TotalEngagement: 750,
		InfluenceScore:  85.5,
	}

	suite.mockService.On("GetUserStats", mock.Anything, "user-456").Return(expectedStats, nil)

	req := httptest.NewRequest("GET", "/api/v1/community/users/user-456/stats", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response UserStatsResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "user-456", response.UserID)
	assert.Equal(suite.T(), 100, response.FollowersCount)
}

func (suite *CommunityHandlerTestSuite) TestGetUserStatsNotFound() {
	suite.mockService.On("GetUserStats", mock.Anything, "nonexistent").Return(nil, ErrUserNotFound)

	req := httptest.NewRequest("GET", "/api/v1/community/users/nonexistent/stats", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test GenerateRecommendations endpoint
func (suite *CommunityHandlerTestSuite) TestGenerateRecommendations() {
	suite.mockService.On("GenerateRecommendations", mock.Anything, "test-user-123", "hybrid").Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/community/recommendations/generate", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Recommendations generated successfully", response["message"])
}

func (suite *CommunityHandlerTestSuite) TestGenerateRecommendationsWithAlgorithm() {
	suite.mockService.On("GenerateRecommendations", mock.Anything, "test-user-123", "collaborative").Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/community/recommendations/generate?algorithm=collaborative", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CommunityHandlerTestSuite) TestGenerateRecommendationsInsufficientData() {
	suite.mockService.On("GenerateRecommendations", mock.Anything, "test-user-123", "hybrid").Return(ErrInsufficientData)

	req := httptest.NewRequest("POST", "/api/v1/community/recommendations/generate", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test CalculateTrending endpoint
func (suite *CommunityHandlerTestSuite) TestCalculateTrending() {
	suite.mockService.On("CalculateTrendingScores", mock.Anything, "daily").Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/community/trending/calculate", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "Trending scores calculated successfully", response["message"])
}

func (suite *CommunityHandlerTestSuite) TestCalculateTrendingWithTimeWindow() {
	suite.mockService.On("CalculateTrendingScores", mock.Anything, "weekly").Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/community/trending/calculate?time_window=weekly", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CommunityHandlerTestSuite) TestCalculateTrendingInvalidTimeWindow() {
	suite.mockService.On("CalculateTrendingScores", mock.Anything, "invalid").Return(ErrInvalidTimeWindow)

	req := httptest.NewRequest("POST", "/api/v1/community/trending/calculate?time_window=invalid", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Test authentication middleware
func (suite *CommunityHandlerTestSuite) TestTrackBehaviorNoAuth() {
	// Create router without auth middleware
	router := gin.New()
	api := router.Group("/api/v1")
	suite.handler.RegisterRoutes(api)

	request := BehaviorTrackingRequest{
		BookmarkID: 1,
		ActionType: "view",
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/community/behavior", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// Run the test suite
func TestCommunityHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CommunityHandlerTestSuite))
}

// Additional unit tests for handler creation
func TestNewHandler(t *testing.T) {
	service := &Service{}
	handler := NewHandler(service)

	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.service)
}
