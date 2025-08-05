package customization

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Simple unit tests for validation and basic functionality
func TestServiceCreation(t *testing.T) {
	service := NewService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestSimpleThemeValidation(t *testing.T) {
	tests := []struct {
		name    string
		theme   Theme
		wantErr bool
	}{
		{
			name: "Valid theme",
			theme: Theme{
				Name:        "test-theme",
				DisplayName: "Test Theme",
				Description: "A test theme",
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			theme: Theme{
				Name:        "",
				DisplayName: "Test Theme",
			},
			wantErr: true,
		},
		{
			name: "Empty display name",
			theme: Theme{
				Name:        "test-theme",
				DisplayName: "",
			},
			wantErr: true,
		},
		{
			name: "Name too short",
			theme: Theme{
				Name:        "ab",
				DisplayName: "Test Theme",
			},
			wantErr: true,
		},
		{
			name: "Name too long",
			theme: Theme{
				Name:        "this-is-a-very-long-theme-name-that-exceeds-the-maximum-allowed-length-of-fifty-characters",
				DisplayName: "Test Theme",
			},
			wantErr: true,
		},
		{
			name: "Description too long",
			theme: Theme{
				Name:        "test-theme",
				DisplayName: "Test Theme",
				Description: "This is a very long description that exceeds the maximum allowed length of 500 characters. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium.",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.theme.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSimpleUserPreferencesValidation(t *testing.T) {
	tests := []struct {
		name    string
		prefs   UserPreferences
		wantErr bool
	}{
		{
			name: "Valid preferences",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			},
			wantErr: false,
		},
		{
			name: "Empty user ID",
			prefs: UserPreferences{
				UserID: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid language",
			prefs: UserPreferences{
				UserID:   "user-123",
				Language: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Invalid grid size",
			prefs: UserPreferences{
				UserID:   "user-123",
				Language: "en",
				GridSize: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Invalid view mode",
			prefs: UserPreferences{
				UserID:   "user-123",
				Language: "en",
				GridSize: "medium",
				ViewMode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Invalid sort by",
			prefs: UserPreferences{
				UserID:    "user-123",
				Language:  "en",
				GridSize:  "medium",
				ViewMode:  "grid",
				SortBy:    "invalid",
				SortOrder: "desc",
			},
			wantErr: true,
		},
		{
			name: "Invalid sort order",
			prefs: UserPreferences{
				UserID:    "user-123",
				Language:  "en",
				GridSize:  "medium",
				ViewMode:  "grid",
				SortBy:    "created_at",
				SortOrder: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Sync interval too low",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 30, // Too low
				SidebarWidth: 250,
			},
			wantErr: true,
		},
		{
			name: "Sync interval too high",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 4000, // Too high
				SidebarWidth: 250,
			},
			wantErr: true,
		},
		{
			name: "Sidebar width too low",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 100, // Too low
			},
			wantErr: true,
		},
		{
			name: "Sidebar width too high",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 600, // Too high
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prefs.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSimpleThemeRatingValidation(t *testing.T) {
	tests := []struct {
		name    string
		rating  ThemeRating
		wantErr bool
	}{
		{
			name: "Valid rating",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  5,
				Comment: "Great theme!",
			},
			wantErr: false,
		},
		{
			name: "Empty user ID",
			rating: ThemeRating{
				UserID: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid theme ID",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 0,
			},
			wantErr: true,
		},
		{
			name: "Rating too low",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  0,
			},
			wantErr: true,
		},
		{
			name: "Rating too high",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  6,
			},
			wantErr: true,
		},
		{
			name: "Comment too long",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  5,
				Comment: "This is a very long comment that exceeds the maximum allowed length of 500 characters. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rating.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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
			err:     ErrInvalidThemeName,
			code:    CodeValidationError,
			message: "Invalid theme name provided",
			expected: ErrorResponse{
				Error:   "invalid theme name",
				Code:    "VALIDATION_ERROR",
				Message: "Invalid theme name provided",
			},
		},
		{
			name:    "Not found error",
			err:     ErrThemeNotFound,
			code:    CodeNotFound,
			message: "Theme not found",
			expected: ErrorResponse{
				Error:   "theme not found",
				Code:    "NOT_FOUND",
				Message: "Theme not found",
			},
		},
		{
			name:    "Permission denied error",
			err:     ErrUnauthorizedTheme,
			code:    CodePermissionDenied,
			message: "Unauthorized to access theme",
			expected: ErrorResponse{
				Error:   "unauthorized to access theme",
				Code:    "PERMISSION_DENIED",
				Message: "Unauthorized to access theme",
			},
		},
		{
			name:    "Already exists error",
			err:     ErrThemeAlreadyExists,
			code:    CodeAlreadyExists,
			message: "Theme with this name already exists",
			expected: ErrorResponse{
				Error:   "theme already exists",
				Code:    "ALREADY_EXISTS",
				Message: "Theme with this name already exists",
			},
		},
		{
			name:    "Internal server error",
			err:     ErrInternalError,
			code:    CodeInternalError,
			message: "An internal server error occurred",
			expected: ErrorResponse{
				Error:   "internal server error",
				Code:    "INTERNAL_ERROR",
				Message: "An internal server error occurred",
			},
		},
		{
			name:    "Unauthorized error",
			err:     ErrPermissionDenied,
			code:    CodeUnauthorized,
			message: "Authentication required",
			expected: ErrorResponse{
				Error:   "permission denied",
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		},
		{
			name:    "Invalid request error",
			err:     ErrInvalidRequest,
			code:    CodeValidationError,
			message: "Request validation failed",
			expected: ErrorResponse{
				Error:   "invalid request",
				Code:    "VALIDATION_ERROR",
				Message: "Request validation failed",
			},
		},
		{
			name:    "Rating already exists error",
			err:     ErrAlreadyRated,
			code:    CodeAlreadyExists,
			message: "User has already rated this theme",
			expected: ErrorResponse{
				Error:   "user has already rated this theme",
				Code:    "ALREADY_EXISTS",
				Message: "User has already rated this theme",
			},
		},
		{
			name:    "Preferences not found error",
			err:     ErrPreferencesNotFound,
			code:    CodeNotFound,
			message: "User preferences not found",
			expected: ErrorResponse{
				Error:   "user preferences not found",
				Code:    "NOT_FOUND",
				Message: "User preferences not found",
			},
		},
		{
			name:    "Rating not found error",
			err:     ErrRatingNotFound,
			code:    CodeNotFound,
			message: "Theme rating not found",
			expected: ErrorResponse{
				Error:   "rating not found",
				Code:    "NOT_FOUND",
				Message: "Theme rating not found",
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

// Test error mapping and unknown error handling - TDD approach
func TestErrorMappingAndUnknownErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected struct {
			code    string
			message string
		}
	}{
		{
			name: "Theme validation error maps to validation code",
			err:  ErrInvalidThemeName,
			expected: struct {
				code    string
				message string
			}{
				code:    CodeValidationError,
				message: "Theme validation failed",
			},
		},
		{
			name: "Theme not found maps to not found code",
			err:  ErrThemeNotFound,
			expected: struct {
				code    string
				message string
			}{
				code:    CodeNotFound,
				message: "Requested theme not found",
			},
		},
		{
			name: "Permission error maps to permission denied code",
			err:  ErrUnauthorizedTheme,
			expected: struct {
				code    string
				message string
			}{
				code:    CodePermissionDenied,
				message: "Access to theme denied",
			},
		},
		{
			name: "Internal error maps to internal error code",
			err:  ErrInternalError,
			expected: struct {
				code    string
				message string
			}{
				code:    CodeInternalError,
				message: "Internal server error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, message := MapErrorToCodeAndMessage(tt.err)
			assert.Equal(t, tt.expected.code, code)
			assert.Equal(t, tt.expected.message, message)
		})
	}
}

// Test unknown error handling - TDD: Write failing test first
func TestUnknownErrorHandling(t *testing.T) {
	// Create an unknown error not defined in our error constants
	unknownErr := errors.New("database connection timeout")

	code, message := MapErrorToCodeAndMessage(unknownErr)

	// Unknown errors should map to internal error
	assert.Equal(t, CodeInternalError, code)
	assert.Equal(t, "An unexpected error occurred", message)

	// Test error response creation for unknown errors
	response := NewErrorResponse(unknownErr, code, message)
	assert.Equal(t, "database connection timeout", response.Error)
	assert.Equal(t, CodeInternalError, response.Code)
	assert.Equal(t, "An unexpected error occurred", response.Message)
}

// Test automatic error response creation - TDD: Write failing test first
func TestAutoErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorResponse
	}{
		{
			name: "Theme validation error auto-mapped",
			err:  ErrInvalidThemeName,
			expected: ErrorResponse{
				Error:   "invalid theme name",
				Code:    CodeValidationError,
				Message: "Theme validation failed",
			},
		},
		{
			name: "Theme not found auto-mapped",
			err:  ErrThemeNotFound,
			expected: ErrorResponse{
				Error:   "theme not found",
				Code:    CodeNotFound,
				Message: "Requested theme not found",
			},
		},
		{
			name: "Permission denied auto-mapped",
			err:  ErrUnauthorizedTheme,
			expected: ErrorResponse{
				Error:   "unauthorized to access theme",
				Code:    CodePermissionDenied,
				Message: "Access to theme denied",
			},
		},
		{
			name: "Unknown error auto-mapped to internal error",
			err:  errors.New("unexpected database error"),
			expected: ErrorResponse{
				Error:   "unexpected database error",
				Code:    CodeInternalError,
				Message: "An unexpected error occurred",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewAutoErrorResponse(tt.err)
			assert.Equal(t, tt.expected.Error, response.Error)
			assert.Equal(t, tt.expected.Code, response.Code)
			assert.Equal(t, tt.expected.Message, response.Message)
		})
	}
}

// Test error response consistency and edge cases - TDD approach
func TestErrorResponseConsistencyAndEdgeCases(t *testing.T) {
	t.Run("Nil error handling", func(t *testing.T) {
		// This should not panic and should handle gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewAutoErrorResponse panicked with nil error: %v", r)
			}
		}()

		// Test with nil error - should be handled gracefully
		response := NewAutoErrorResponse(nil)
		assert.Equal(t, CodeInternalError, response.Code)
		assert.Equal(t, "An unexpected error occurred", response.Message)
	})

	t.Run("Error response JSON serialization", func(t *testing.T) {
		response := NewAutoErrorResponse(ErrThemeNotFound)

		// Verify all fields are properly set for JSON serialization
		assert.NotEmpty(t, response.Error)
		assert.NotEmpty(t, response.Code)
		assert.NotEmpty(t, response.Message)

		// Verify specific values
		assert.Equal(t, "theme not found", response.Error)
		assert.Equal(t, CodeNotFound, response.Code)
		assert.Equal(t, "Requested theme not found", response.Message)
	})

	t.Run("All error codes are covered", func(t *testing.T) {
		// Test that all defined error codes have corresponding error mappings
		errorCodeTests := []struct {
			err  error
			code string
		}{
			{ErrInvalidThemeName, CodeValidationError},
			{ErrThemeNotFound, CodeNotFound},
			{ErrThemeAlreadyExists, CodeAlreadyExists},
			{ErrUnauthorizedTheme, CodePermissionDenied},
			{ErrInternalError, CodeInternalError},
		}

		for _, test := range errorCodeTests {
			code, _ := MapErrorToCodeAndMessage(test.err)
			assert.Equal(t, test.code, code, "Error %v should map to code %s", test.err, test.code)
		}
	})

	t.Run("Error message consistency", func(t *testing.T) {
		// Test that similar errors get consistent message patterns
		validationErrors := []error{
			ErrInvalidThemeName,
			ErrInvalidDisplayName,
			ErrInvalidThemeConfig,
		}

		for _, err := range validationErrors {
			code, message := MapErrorToCodeAndMessage(err)
			assert.Equal(t, CodeValidationError, code)
			assert.Contains(t, message, "validation failed", "Validation errors should have consistent message pattern")
		}

		notFoundErrors := []error{
			ErrThemeNotFound,
			ErrPreferencesNotFound,
			ErrRatingNotFound,
		}

		for _, err := range notFoundErrors {
			code, message := MapErrorToCodeAndMessage(err)
			assert.Equal(t, CodeNotFound, code)
			assert.Contains(t, message, "not found", "Not found errors should have consistent message pattern")
		}
	})
}

func TestRequestValidation(t *testing.T) {
	// Test CreateThemeRequest validation
	t.Run("CreateThemeRequest", func(t *testing.T) {
		validReq := CreateThemeRequest{
			Name:        "test-theme",
			DisplayName: "Test Theme",
			Description: "A test theme",
			IsPublic:    true,
			Config:      map[string]any{"color": "blue"},
		}

		// This would be validated by Gin's binding, but we can test the structure
		assert.Equal(t, "test-theme", validReq.Name)
		assert.Equal(t, "Test Theme", validReq.DisplayName)
		assert.True(t, validReq.IsPublic)
	})

	// Test UpdateUserPreferencesRequest validation
	t.Run("UpdateUserPreferencesRequest", func(t *testing.T) {
		language := "zh-CN"
		gridSize := "large"
		showThumbnails := true
		syncInterval := 600

		validReq := UpdateUserPreferencesRequest{
			Language:       language,
			GridSize:       gridSize,
			ShowThumbnails: &showThumbnails,
			SyncInterval:   &syncInterval,
		}

		assert.Equal(t, "zh-CN", validReq.Language)
		assert.Equal(t, "large", validReq.GridSize)
		assert.True(t, *validReq.ShowThumbnails)
		assert.Equal(t, 600, *validReq.SyncInterval)
	})

	// Test RateThemeRequest validation
	t.Run("RateThemeRequest", func(t *testing.T) {
		validReq := RateThemeRequest{
			Rating:  5,
			Comment: "Excellent theme!",
		}

		assert.Equal(t, 5, validReq.Rating)
		assert.Equal(t, "Excellent theme!", validReq.Comment)
	})
}

func TestLanguageSupport(t *testing.T) {
	supportedLanguages := []string{"en", "zh-CN", "zh-TW", "ja", "ko"}

	for _, lang := range supportedLanguages {
		t.Run("Language_"+lang, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     lang,
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.NoError(t, err, "Language %s should be supported", lang)
		})
	}

	// Test unsupported language
	t.Run("Unsupported_Language", func(t *testing.T) {
		prefs := UserPreferences{
			UserID:   "user-123",
			Language: "unsupported",
		}

		err := prefs.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidLanguage, err)
	})
}

func TestGridSizeOptions(t *testing.T) {
	validSizes := []string{"small", "medium", "large"}

	for _, size := range validSizes {
		t.Run("GridSize_"+size, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     size,
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.NoError(t, err, "Grid size %s should be valid", size)
		})
	}
}

// Test invalid grid size options - TDD: Write failing test first
func TestInvalidGridSizeOptions(t *testing.T) {
	invalidSizes := []string{
		"tiny",         // Not in valid list
		"extra-large",  // Not in valid list
		"xl",           // Not in valid list
		"SMALL",        // Case sensitive - should be lowercase
		"Medium",       // Case sensitive - should be lowercase
		"LARGE",        // Case sensitive - should be lowercase
		"",             // Empty string (if GridSize is required)
		"invalid",      // Generic invalid value
		"mini",         // Another invalid size
		"huge",         // Another invalid size
		"1",            // Numeric value
		"small-medium", // Hyphenated invalid value
	}

	for _, size := range invalidSizes {
		t.Run("InvalidGridSize_"+size, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     size,
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.Error(t, err, "Grid size '%s' should be invalid", size)
			assert.Equal(t, ErrInvalidGridSize, err, "Should return ErrInvalidGridSize for '%s'", size)
		})
	}
}

func TestViewModeOptions(t *testing.T) {
	validModes := []string{"grid", "list", "compact"}

	for _, mode := range validModes {
		t.Run("ViewMode_"+mode, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     mode,
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.NoError(t, err, "View mode %s should be valid", mode)
		})
	}
}

// Test invalid view mode options - TDD: Write failing test first
func TestInvalidViewModeOptions(t *testing.T) {
	invalidModes := []string{
		"table",     // Not in valid list
		"card",      // Not in valid list
		"tile",      // Not in valid list
		"GRID",      // Case sensitive - should be lowercase
		"List",      // Case sensitive - should be lowercase
		"COMPACT",   // Case sensitive - should be lowercase
		"",          // Empty string
		"invalid",   // Generic invalid value
		"gallery",   // Another invalid mode
		"thumbnail", // Another invalid mode
		"1",         // Numeric value
		"grid-view", // Hyphenated invalid value
		"list_view", // Underscore invalid value
	}

	for _, mode := range invalidModes {
		t.Run("InvalidViewMode_"+mode, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     mode,
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.Error(t, err, "View mode '%s' should be invalid", mode)
			assert.Equal(t, ErrInvalidViewMode, err, "Should return ErrInvalidViewMode for '%s'", mode)
		})
	}
}

// Test invalid sort by options - TDD: Write failing test first
func TestInvalidSortByOptions(t *testing.T) {
	invalidSortBy := []string{
		"name",       // Not in valid list
		"date",       // Not in valid list
		"modified",   // Not in valid list
		"CREATED_AT", // Case sensitive - should be lowercase
		"Updated_At", // Case sensitive - should be lowercase
		"TITLE",      // Case sensitive - should be lowercase
		"",           // Empty string
		"invalid",    // Generic invalid value
		"popularity", // Another invalid sort option
		"rating",     // Another invalid sort option
		"1",          // Numeric value
		"created-at", // Hyphenated invalid value
		"updated.at", // Dot notation invalid value
	}

	for _, sortBy := range invalidSortBy {
		t.Run("InvalidSortBy_"+sortBy, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       sortBy,
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.Error(t, err, "Sort by '%s' should be invalid", sortBy)
			assert.Equal(t, ErrInvalidSortBy, err, "Should return ErrInvalidSortBy for '%s'", sortBy)
		})
	}
}

// Test invalid sort order options - TDD: Write failing test first
func TestInvalidSortOrderOptions(t *testing.T) {
	invalidSortOrder := []string{
		"ascending",  // Not in valid list
		"descending", // Not in valid list
		"up",         // Not in valid list
		"down",       // Not in valid list
		"ASC",        // Case sensitive - should be lowercase
		"DESC",       // Case sensitive - should be lowercase
		"Asc",        // Case sensitive - should be lowercase
		"Desc",       // Case sensitive - should be lowercase
		"",           // Empty string
		"invalid",    // Generic invalid value
		"1",          // Numeric value
		"0",          // Numeric value
		"true",       // Boolean as string
		"false",      // Boolean as string
	}

	for _, sortOrder := range invalidSortOrder {
		t.Run("InvalidSortOrder_"+sortOrder, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    sortOrder,
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.Error(t, err, "Sort order '%s' should be invalid", sortOrder)
			assert.Equal(t, ErrInvalidSortOrder, err, "Should return ErrInvalidSortOrder for '%s'", sortOrder)
		})
	}
}

// Test valid sort by options - TDD: Ensure positive cases still work
func TestValidSortByOptions(t *testing.T) {
	validSortBy := []string{"created_at", "updated_at", "title", "url"}

	for _, sortBy := range validSortBy {
		t.Run("ValidSortBy_"+sortBy, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       sortBy,
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.NoError(t, err, "Sort by '%s' should be valid", sortBy)
		})
	}
}

// Test valid sort order options - TDD: Ensure positive cases still work
func TestValidSortOrderOptions(t *testing.T) {
	validSortOrder := []string{"asc", "desc"}

	for _, sortOrder := range validSortOrder {
		t.Run("ValidSortOrder_"+sortOrder, func(t *testing.T) {
			prefs := UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    sortOrder,
				SyncInterval: 300,
				SidebarWidth: 250,
			}

			err := prefs.Validate()
			assert.NoError(t, err, "Sort order '%s' should be valid", sortOrder)
		})
	}
}
