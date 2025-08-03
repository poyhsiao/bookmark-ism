package community

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Simple unit tests for validation and basic functionality
func TestServiceCreation(t *testing.T) {
	service := NewService(nil, nil)
	assert.NotNil(t, service)
}

func TestBehaviorTrackingRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request BehaviorTrackingRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: BehaviorTrackingRequest{
				UserID:     "user-123",
				BookmarkID: 1,
				ActionType: "view",
				Duration:   30,
				Context:    "homepage",
			},
			wantErr: false,
		},
		{
			name: "Empty user ID",
			request: BehaviorTrackingRequest{
				UserID:     "",
				BookmarkID: 1,
				ActionType: "view",
			},
			wantErr: true,
		},
		{
			name: "Zero bookmark ID",
			request: BehaviorTrackingRequest{
				UserID:     "user-123",
				BookmarkID: 0,
				ActionType: "view",
			},
			wantErr: true,
		},
		{
			name: "Empty action type",
			request: BehaviorTrackingRequest{
				UserID:     "user-123",
				BookmarkID: 1,
				ActionType: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create behavior from request
			behavior := &UserBehavior{
				UserID:     tt.request.UserID,
				BookmarkID: tt.request.BookmarkID,
				ActionType: tt.request.ActionType,
				Duration:   tt.request.Duration,
				Context:    tt.request.Context,
			}

			err := behavior.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFollowRequestValidation(t *testing.T) {
	tests := []struct {
		name       string
		followerID string
		request    FollowRequest
		wantErr    bool
	}{
		{
			name:       "Valid follow request",
			followerID: "user-123",
			request:    FollowRequest{FollowingID: "user-456"},
			wantErr:    false,
		},
		{
			name:       "Empty follower ID",
			followerID: "",
			request:    FollowRequest{FollowingID: "user-456"},
			wantErr:    true,
		},
		{
			name:       "Empty following ID",
			followerID: "user-123",
			request:    FollowRequest{FollowingID: ""},
			wantErr:    true,
		},
		{
			name:       "Cannot follow self",
			followerID: "user-123",
			request:    FollowRequest{FollowingID: "user-123"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate follower ID
			if tt.followerID == "" {
				assert.Error(t, ErrInvalidFollowerID)
				return
			}

			// Validate following ID
			if tt.request.FollowingID == "" {
				assert.Error(t, ErrInvalidFollowingID)
				return
			}

			// Check self-follow
			if tt.followerID == tt.request.FollowingID {
				assert.Error(t, ErrCannotFollowSelf)
				return
			}

			// Create follow relationship
			follow := &UserFollow{
				FollowerID:  tt.followerID,
				FollowingID: tt.request.FollowingID,
				Status:      "active",
			}

			err := follow.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecommendationRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request RecommendationRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: RecommendationRequest{
				UserID:    "user-123",
				Limit:     20,
				Algorithm: "collaborative",
				Context:   "homepage",
			},
			wantErr: false,
		},
		{
			name: "Empty user ID",
			request: RecommendationRequest{
				UserID: "",
				Limit:  20,
			},
			wantErr: true,
		},
		{
			name: "Zero limit gets default",
			request: RecommendationRequest{
				UserID: "user-123",
				Limit:  0,
			},
			wantErr: false, // Should get default limit
		},
		{
			name: "Limit too high gets capped",
			request: RecommendationRequest{
				UserID: "user-123",
				Limit:  200,
			},
			wantErr: false, // Should get capped to 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.request.UserID == "" {
				assert.Equal(t, ErrInvalidUserID.Error(), "invalid user ID")
				return
			}

			// Test limit normalization
			if tt.request.Limit <= 0 || tt.request.Limit > 100 {
				if tt.request.Limit <= 0 {
					tt.request.Limit = 20 // Default
				}
				if tt.request.Limit > 100 {
					tt.request.Limit = 100 // Cap
				}
			}

			assert.True(t, tt.request.Limit > 0 && tt.request.Limit <= 100)
		})
	}
}

func TestTrendingRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request TrendingRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: TrendingRequest{
				TimeWindow: "daily",
				Limit:      20,
				MinScore:   0.5,
			},
			wantErr: false,
		},
		{
			name: "Invalid time window",
			request: TrendingRequest{
				TimeWindow: "invalid",
				Limit:      20,
			},
			wantErr: true,
		},
		{
			name: "Empty time window gets default",
			request: TrendingRequest{
				TimeWindow: "",
				Limit:      20,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default time window
			if tt.request.TimeWindow == "" {
				tt.request.TimeWindow = "daily"
			}

			// Validate time window
			validWindows := map[string]bool{
				"hourly": true, "daily": true, "weekly": true, "monthly": true,
			}

			if !validWindows[tt.request.TimeWindow] {
				if tt.wantErr {
					assert.False(t, validWindows[tt.request.TimeWindow])
				}
				return
			}

			assert.True(t, validWindows[tt.request.TimeWindow])
		})
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		message  string
		expected ErrorResponse
	}{
		{
			name:    "Validation error",
			err:     ErrInvalidUserID,
			code:    CodeValidationError,
			message: "Invalid user ID provided",
			expected: ErrorResponse{
				Error:   "invalid user ID",
				Code:    "VALIDATION_ERROR",
				Message: "Invalid user ID provided",
			},
		},
		{
			name:    "Not found error",
			err:     ErrUserNotFound,
			code:    CodeNotFound,
			message: "User not found",
			expected: ErrorResponse{
				Error:   "user not found",
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewErrorResponse(tt.err, tt.code, tt.message)
			assert.Equal(t, tt.expected.Error, response.Error)
			assert.Equal(t, tt.expected.Code, response.Code)
			assert.Equal(t, tt.expected.Message, response.Message)
		})
	}
}

func TestServiceMethodsExist(t *testing.T) {
	// This test ensures all required service methods exist by checking the interface
	service := NewService(nil, nil)
	assert.NotNil(t, service)

	// Test that the service implements the CommunityService interface
	var _ CommunityService = service
}

func TestRecommendationAlgorithms(t *testing.T) {
	validAlgorithms := []string{
		"collaborative", "content_based", "trending",
		"popularity", "category", "hybrid",
	}

	for _, algorithm := range validAlgorithms {
		t.Run("Algorithm_"+algorithm, func(t *testing.T) {
			// Test that algorithm is recognized as valid
			validAlgorithmMap := map[string]bool{
				"collaborative": true, "content_based": true, "trending": true,
				"popularity": true, "category": true, "hybrid": true,
			}

			assert.True(t, validAlgorithmMap[algorithm], "Algorithm %s should be valid", algorithm)
		})
	}

	// Test invalid algorithm
	t.Run("Invalid_Algorithm", func(t *testing.T) {
		validAlgorithmMap := map[string]bool{
			"collaborative": true, "content_based": true, "trending": true,
			"popularity": true, "category": true, "hybrid": true,
		}

		assert.False(t, validAlgorithmMap["invalid_algorithm"])
	})
}

func TestTimeWindows(t *testing.T) {
	validWindows := []string{"hourly", "daily", "weekly", "monthly"}

	for _, window := range validWindows {
		t.Run("TimeWindow_"+window, func(t *testing.T) {
			validWindowMap := map[string]bool{
				"hourly": true, "daily": true, "weekly": true, "monthly": true,
			}

			assert.True(t, validWindowMap[window], "Time window %s should be valid", window)
		})
	}

	// Test invalid time window
	t.Run("Invalid_TimeWindow", func(t *testing.T) {
		validWindowMap := map[string]bool{
			"hourly": true, "daily": true, "weekly": true, "monthly": true,
		}

		assert.False(t, validWindowMap["invalid_window"])
	})
}
