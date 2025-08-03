package customization

import "errors"

// Error definitions
var (
	// Theme errors
	ErrThemeNotFound      = errors.New("theme not found")
	ErrThemeAlreadyExists = errors.New("theme already exists")
	ErrInvalidThemeName   = errors.New("invalid theme name")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidThemeConfig = errors.New("invalid theme configuration")
	ErrThemeNotPublic     = errors.New("theme is not public")
	ErrUnauthorizedTheme  = errors.New("unauthorized to access theme")

	// User preferences errors
	ErrPreferencesNotFound = errors.New("user preferences not found")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidLanguage     = errors.New("invalid language")
	ErrInvalidTimezone     = errors.New("invalid timezone")
	ErrInvalidDateFormat   = errors.New("invalid date format")
	ErrInvalidTimeFormat   = errors.New("invalid time format")
	ErrInvalidGridSize     = errors.New("invalid grid size")
	ErrInvalidViewMode     = errors.New("invalid view mode")
	ErrInvalidSortBy       = errors.New("invalid sort by field")
	ErrInvalidSortOrder    = errors.New("invalid sort order")
	ErrInvalidSyncInterval = errors.New("invalid sync interval")
	ErrInvalidSidebarWidth = errors.New("invalid sidebar width")

	// Theme rating errors
	ErrInvalidThemeID = errors.New("invalid theme ID")
	ErrInvalidRating  = errors.New("invalid rating")
	ErrInvalidComment = errors.New("invalid comment")
	ErrRatingNotFound = errors.New("rating not found")
	ErrAlreadyRated   = errors.New("user has already rated this theme")

	// General errors
	ErrInvalidRequest   = errors.New("invalid request")
	ErrInternalError    = errors.New("internal server error")
	ErrPermissionDenied = errors.New("permission denied")
)

// Error codes for API responses
const (
	CodeValidationError  = "VALIDATION_ERROR"
	CodeNotFound         = "NOT_FOUND"
	CodeAlreadyExists    = "ALREADY_EXISTS"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeInternalError    = "INTERNAL_ERROR"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, code, message string) ErrorResponse {
	return ErrorResponse{
		Error:   err.Error(),
		Code:    code,
		Message: message,
	}
}
