package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PreferenceValidatorTestSuite struct {
	suite.Suite
	validator *PreferenceValidator
}

func (suite *PreferenceValidatorTestSuite) SetupTest() {
	suite.validator = NewPreferenceValidator()
}

// TestValidateTheme tests theme validation
func (suite *PreferenceValidatorTestSuite) TestValidateTheme() {
	testCases := []struct {
		name        string
		theme       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid theme - light",
			theme:       "light",
			expectError: false,
		},
		{
			name:        "Valid theme - dark",
			theme:       "dark",
			expectError: false,
		},
		{
			name:        "Valid theme - auto",
			theme:       "auto",
			expectError: false,
		},
		{
			name:        "Invalid theme - neon",
			theme:       "neon",
			expectError: true,
			errorMsg:    "invalid theme 'neon', must be one of: light, dark, auto",
		},
		{
			name:        "Invalid theme - blue",
			theme:       "blue",
			expectError: true,
			errorMsg:    "invalid theme 'blue', must be one of: light, dark, auto",
		},
		{
			name:        "Invalid theme - empty string",
			theme:       "",
			expectError: false, // Empty strings are handled at request level
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.validator.validateTheme(tc.theme)
			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidateGridSize tests grid size validation
func (suite *PreferenceValidatorTestSuite) TestValidateGridSize() {
	testCases := []struct {
		name        string
		gridSize    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid grid size - small",
			gridSize:    "small",
			expectError: false,
		},
		{
			name:        "Valid grid size - medium",
			gridSize:    "medium",
			expectError: false,
		},
		{
			name:        "Valid grid size - large",
			gridSize:    "large",
			expectError: false,
		},
		{
			name:        "Invalid grid size - tiny",
			gridSize:    "tiny",
			expectError: true,
			errorMsg:    "invalid gridSize 'tiny', must be one of: small, medium, large",
		},
		{
			name:        "Invalid grid size - extra_large",
			gridSize:    "extra_large",
			expectError: true,
			errorMsg:    "invalid gridSize 'extra_large', must be one of: small, medium, large",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.validator.validateGridSize(tc.gridSize)
			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidateDefaultView tests default view validation
func (suite *PreferenceValidatorTestSuite) TestValidateDefaultView() {
	testCases := []struct {
		name        string
		defaultView string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid default view - grid",
			defaultView: "grid",
			expectError: false,
		},
		{
			name:        "Valid default view - list",
			defaultView: "list",
			expectError: false,
		},
		{
			name:        "Invalid default view - card",
			defaultView: "card",
			expectError: true,
			errorMsg:    "invalid defaultView 'card', must be one of: grid, list",
		},
		{
			name:        "Invalid default view - carousel",
			defaultView: "carousel",
			expectError: true,
			errorMsg:    "invalid defaultView 'carousel', must be one of: grid, list",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.validator.validateDefaultView(tc.defaultView)
			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidateLanguage tests language validation
func (suite *PreferenceValidatorTestSuite) TestValidateLanguage() {
	testCases := []struct {
		name        string
		language    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid language - en",
			language:    "en",
			expectError: false,
		},
		{
			name:        "Valid language - zh-CN",
			language:    "zh-CN",
			expectError: false,
		},
		{
			name:        "Valid language - zh-TW",
			language:    "zh-TW",
			expectError: false,
		},
		{
			name:        "Invalid language - fr",
			language:    "fr",
			expectError: true,
			errorMsg:    "invalid language 'fr', must be one of: en, zh-CN, zh-TW",
		},
		{
			name:        "Invalid language - klingon",
			language:    "klingon",
			expectError: true,
			errorMsg:    "invalid language 'klingon', must be one of: en, zh-CN, zh-TW",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.validator.validateLanguage(tc.language)
			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidateTimezone tests timezone validation
func (suite *PreferenceValidatorTestSuite) TestValidateTimezone() {
	testCases := []struct {
		name        string
		timezone    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid timezone - UTC",
			timezone:    "UTC",
			expectError: false,
		},
		{
			name:        "Valid timezone - America/New_York",
			timezone:    "America/New_York",
			expectError: false,
		},
		{
			name:        "Valid timezone - Asia/Shanghai",
			timezone:    "Asia/Shanghai",
			expectError: false,
		},
		{
			name:        "Valid timezone - Europe/London",
			timezone:    "Europe/London",
			expectError: false,
		},
		{
			name:        "Invalid timezone - Invalid/Timezone",
			timezone:    "Invalid/Timezone",
			expectError: true,
			errorMsg:    "invalid timezone 'Invalid/Timezone'",
		},
		{
			name:        "Invalid timezone - NotATimezone",
			timezone:    "NotATimezone",
			expectError: true,
			errorMsg:    "invalid timezone 'NotATimezone'",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.validator.validateTimezone(tc.timezone)
			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidatePreferences tests the complete preferences validation
func (suite *PreferenceValidatorTestSuite) TestValidatePreferences() {
	testCases := []struct {
		name        string
		request     *UpdatePreferencesRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid preferences - all fields",
			request: &UpdatePreferencesRequest{
				Theme:       "dark",
				GridSize:    "large",
				DefaultView: "list",
				Language:    "zh-CN",
				Timezone:    "Asia/Shanghai",
			},
			expectError: false,
		},
		{
			name: "Valid preferences - partial fields",
			request: &UpdatePreferencesRequest{
				Theme:    "auto",
				Language: "zh-TW",
			},
			expectError: false,
		},
		{
			name:        "Valid preferences - empty request",
			request:     &UpdatePreferencesRequest{},
			expectError: false,
		},
		{
			name: "Invalid preferences - single field",
			request: &UpdatePreferencesRequest{
				Theme: "neon",
			},
			expectError: true,
			errorMsg:    "validation failed: invalid theme 'neon'",
		},
		{
			name: "Invalid preferences - multiple fields",
			request: &UpdatePreferencesRequest{
				Theme:       "neon",
				GridSize:    "tiny",
				DefaultView: "carousel",
				Language:    "klingon",
				Timezone:    "Invalid/Timezone",
			},
			expectError: true,
			errorMsg:    "validation failed",
		},
		{
			name: "Mixed valid and invalid preferences",
			request: &UpdatePreferencesRequest{
				Theme:       "dark", // valid
				GridSize:    "tiny", // invalid
				DefaultView: "list", // valid
				Language:    "fr",   // invalid
			},
			expectError: true,
			errorMsg:    "validation failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.validator.ValidatePreferences(tc.request)
			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.errorMsg)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestValidatePreferences_MultipleErrors tests that multiple validation errors are properly combined
func (suite *PreferenceValidatorTestSuite) TestValidatePreferences_MultipleErrors() {
	request := &UpdatePreferencesRequest{
		Theme:       "invalid_theme",
		GridSize:    "invalid_size",
		DefaultView: "invalid_view",
		Language:    "invalid_lang",
		Timezone:    "Invalid/Timezone",
	}

	err := suite.validator.ValidatePreferences(request)
	assert.Error(suite.T(), err)

	errorMsg := err.Error()
	assert.Contains(suite.T(), errorMsg, "validation failed")
	assert.Contains(suite.T(), errorMsg, "invalid theme")
	assert.Contains(suite.T(), errorMsg, "invalid gridSize")
	assert.Contains(suite.T(), errorMsg, "invalid defaultView")
	assert.Contains(suite.T(), errorMsg, "invalid language")
	assert.Contains(suite.T(), errorMsg, "invalid timezone")
}

func TestPreferenceValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(PreferenceValidatorTestSuite))
}
