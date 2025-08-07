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
)

// Simple mock service for basic testing
type SimpleMockService struct{}

func (s *SimpleMockService) TrackUserBehavior(ctx context.Context, req *BehaviorTrackingRequest) error {
	// Note: UserID is set by the handler from auth context, so we don't validate it here
	if req.BookmarkID == 0 {
		return ErrInvalidBookmarkID
	}
	if req.ActionType == "" {
		return ErrInvalidActionType
	}
	return nil
}

func (s *SimpleMockService) FollowUser(ctx context.Context, followerID string, req *FollowRequest) error {
	if followerID == "" {
		return ErrInvalidFollowerID
	}
	if req.FollowingID == "" {
		return ErrInvalidFollowingID
	}
	if followerID == req.FollowingID {
		return ErrCannotFollowSelf
	}
	if followerID == "already-following" {
		return ErrAlreadyFollowing
	}
	return nil
}

func (s *SimpleMockService) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	if followerID == "" {
		return ErrInvalidFollowerID
	}
	if followingID == "" {
		return ErrInvalidFollowingID
	}
	if followingID == "not-following" {
		return ErrNotFollowing
	}
	return nil
}

func (s *SimpleMockService) GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]RecommendationResponse, error) {
	if req.UserID == "" {
		return nil, ErrInvalidUserID
	}
	if req.UserID == "no-data" {
		return nil, ErrInsufficientData
	}

	return []RecommendationResponse{
		{
			BookmarkID: 1,
			Score:      0.8,
			ReasonType: "collaborative",
			ReasonText: "Users with similar interests also liked this",
		},
	}, nil
}

func (s *SimpleMockService) GetTrendingBookmarksInternal(ctx context.Context, req *TrendingRequest) ([]TrendingResponse, error) {
	validWindows := map[string]bool{
		"hourly": true, "daily": true, "weekly": true, "monthly": true,
	}
	if !validWindows[req.TimeWindow] {
		return nil, ErrInvalidTimeWindow
	}

	return []TrendingResponse{
		{
			BookmarkID:     1,
			TrendingScore:  95.5,
			ViewCount:      1000,
			EngagementRate: 0.15,
			TimeWindow:     req.TimeWindow,
		},
	}, nil
}

func (s *SimpleMockService) GenerateUserFeed(ctx context.Context, req *FeedRequest) ([]UserFeedResponse, error) {
	if req.UserID == "" {
		return nil, ErrInvalidUserID
	}

	return []UserFeedResponse{
		{
			BookmarkID: 1,
			SourceType: "following",
			SourceID:   "user-456",
			Score:      0.9,
			Position:   1,
		},
	}, nil
}

func (s *SimpleMockService) GetSocialMetrics(ctx context.Context, bookmarkID uint) (*SocialMetricsResponse, error) {
	if bookmarkID == 0 {
		return nil, ErrInvalidBookmarkID
	}
	if bookmarkID == 999 {
		return nil, ErrBookmarkNotFound
	}

	return &SocialMetricsResponse{
		BookmarkID:     bookmarkID,
		TotalViews:     1000,
		TotalClicks:    150,
		TotalSaves:     50,
		TotalShares:    25,
		TotalLikes:     75,
		UniqueViewers:  800,
		EngagementRate: 0.15,
		ViralityScore:  125.0,
		QualityScore:   0.8,
	}, nil
}

func (s *SimpleMockService) GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if userID == "not-found" {
		return nil, ErrUserNotFound
	}

	return &UserStatsResponse{
		UserID:          userID,
		FollowersCount:  100,
		FollowingCount:  50,
		BookmarksCount:  200,
		PublicBookmarks: 150,
		TotalViews:      5000,
		TotalEngagement: 750,
		InfluenceScore:  85.5,
	}, nil
}

func (s *SimpleMockService) GenerateRecommendations(ctx context.Context, userID, algorithm string) error {
	if userID == "" {
		return ErrInvalidUserID
	}
	validAlgorithms := map[string]bool{
		"collaborative": true, "content_based": true, "trending": true,
		"popularity": true, "category": true, "hybrid": true,
	}
	if !validAlgorithms[algorithm] {
		return ErrInvalidAlgorithm
	}
	if userID == "no-data" {
		return ErrInsufficientData
	}
	return nil
}

func (s *SimpleMockService) CalculateTrendingScores(ctx context.Context, timeWindow string) error {
	validWindows := map[string]bool{
		"hourly": true, "daily": true, "weekly": true, "monthly": true,
	}
	if !validWindows[timeWindow] {
		return ErrInvalidTimeWindow
	}
	return nil
}

// Test helper to create router with auth middleware
func setupTestRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)

	service := &SimpleMockService{}
	handler := NewHandler(service)

	router := gin.New()

	// Add auth middleware mock
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})

	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	return router, handler
}

func TestHandlerCreation(t *testing.T) {
	service := &SimpleMockService{}
	handler := NewHandler(service)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.service)
}

func TestTrackBehaviorHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		request        BehaviorTrackingRequest
		expectedStatus int
	}{
		{
			name: "Valid request",
			request: BehaviorTrackingRequest{
				BookmarkID: 1,
				ActionType: "view",
				Duration:   30,
				Context:    "homepage",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid bookmark ID",
			request: BehaviorTrackingRequest{
				BookmarkID: 0,
				ActionType: "view",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid action type",
			request: BehaviorTrackingRequest{
				BookmarkID: 1,
				ActionType: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/community/behavior", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestFollowUserHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		request        FollowRequest
		expectedStatus int
	}{
		{
			name:           "Valid follow request",
			request:        FollowRequest{FollowingID: "user-456"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty following ID",
			request:        FollowRequest{FollowingID: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Cannot follow self",
			request:        FollowRequest{FollowingID: "test-user-123"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/community/follow", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUnfollowUserHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Valid unfollow request",
			userID:         "user-456",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not following user",
			userID:         "not-following",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/follow/" + tt.userID
			req := httptest.NewRequest("DELETE", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetRecommendationsHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With algorithm parameter",
			queryParams:    "?algorithm=collaborative",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With limit parameter",
			queryParams:    "?limit=10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/recommendations" + tt.queryParams
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response, "recommendations")
				assert.Contains(t, response, "total")
			}
		})
	}
}

func TestGetTrendingHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With time window",
			queryParams:    "?time_window=weekly",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid time window",
			queryParams:    "?time_window=invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/trending" + tt.queryParams
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetSocialMetricsHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		bookmarkID     string
		expectedStatus int
	}{
		{
			name:           "Valid bookmark ID",
			bookmarkID:     "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid bookmark ID",
			bookmarkID:     "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bookmark not found",
			bookmarkID:     "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/metrics/" + tt.bookmarkID
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetUserStatsHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Valid user ID",
			userID:         "user-456",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty user ID",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "User not found",
			userID:         "not-found",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/users/" + tt.userID + "/stats"
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGenerateRecommendationsHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With algorithm",
			queryParams:    "?algorithm=collaborative",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/recommendations/generate" + tt.queryParams
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCalculateTrendingHandler(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With time window",
			queryParams:    "?time_window=weekly",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid time window",
			queryParams:    "?time_window=invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/community/trending/calculate" + tt.queryParams
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthenticationRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &SimpleMockService{}
	handler := NewHandler(service)

	// Create router without auth middleware
	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	// Test endpoints that require authentication
	endpoints := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"POST", "/api/v1/community/behavior", BehaviorTrackingRequest{BookmarkID: 1, ActionType: "view"}},
		{"POST", "/api/v1/community/follow", FollowRequest{FollowingID: "user-456"}},
		{"GET", "/api/v1/community/recommendations", nil},
		{"GET", "/api/v1/community/feed", nil},
		{"POST", "/api/v1/community/recommendations/generate", nil},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+"_"+endpoint.path, func(t *testing.T) {
			var req *http.Request
			if endpoint.body != nil {
				body, _ := json.Marshal(endpoint.body)
				req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
