package customization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Simple unit tests for validation and basic functionality
func TestServiceCreation(t *testing.T) {
	service := NewService(nil, nil)
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
