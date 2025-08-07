package community

import "errors"

// Community-specific errors
var (
	// Validation errors
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidBookmarkID   = errors.New("invalid bookmark ID")
	ErrInvalidActionType   = errors.New("invalid action type")
	ErrInvalidFollowerID   = errors.New("invalid follower ID")
	ErrInvalidFollowingID  = errors.New("invalid following ID")
	ErrCannotFollowSelf    = errors.New("cannot follow yourself")
	ErrInvalidFollowStatus = errors.New("invalid follow status")
	ErrInvalidScore        = errors.New("score must be between 0 and 1")
	ErrInvalidReasonType   = errors.New("invalid recommendation reason type")
	ErrInvalidTimeWindow   = errors.New("invalid time window")
	ErrInvalidAlgorithm    = errors.New("invalid recommendation algorithm")

	// Business logic errors
	ErrUserNotFound           = errors.New("user not found")
	ErrBookmarkNotFound       = errors.New("bookmark not found")
	ErrAlreadyFollowing       = errors.New("already following this user")
	ErrNotFollowing           = errors.New("not following this user")
	ErrRecommendationNotFound = errors.New("recommendation not found")
	ErrInsufficientData       = errors.New("insufficient data for recommendations")
	ErrPrivacyRestriction     = errors.New("privacy settings restrict this action")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")

	// System errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrCacheConnection    = errors.New("cache connection error")
	ErrExternalService    = errors.New("external service error")
	ErrInternalServer     = errors.New("internal server error")
)

// Error response structure
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error codes for client handling
const (
	CodeValidationError    = "VALIDATION_ERROR"
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Helper function to create error responses
func NewErrorResponse(err error, code string, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   err.Error(),
		Code:    code,
		Message: message,
	}
}
