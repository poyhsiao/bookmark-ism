package user

import (
	"fmt"
	"strings"
	"time"
)

// PreferenceValidator handles validation of user preferences
type PreferenceValidator struct{}

// NewPreferenceValidator creates a new preference validator
func NewPreferenceValidator() *PreferenceValidator {
	return &PreferenceValidator{}
}

// ValidatePreferences validates the UpdatePreferencesRequest
func (v *PreferenceValidator) ValidatePreferences(req *UpdatePreferencesRequest) error {
	var errors []string

	// Validate theme
	if req.Theme != "" {
		if err := v.validateTheme(req.Theme); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// Validate grid size
	if req.GridSize != "" {
		if err := v.validateGridSize(req.GridSize); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// Validate default view
	if req.DefaultView != "" {
		if err := v.validateDefaultView(req.DefaultView); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// Validate language
	if req.Language != "" {
		if err := v.validateLanguage(req.Language); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// Validate timezone
	if req.Timezone != "" {
		if err := v.validateTimezone(req.Timezone); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateTheme validates the theme preference
func (v *PreferenceValidator) validateTheme(theme string) error {
	if theme == "" {
		return nil // Empty strings are handled at request level
	}

	validThemes := []string{"light", "dark", "auto"}
	for _, valid := range validThemes {
		if theme == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid theme '%s', must be one of: %s", theme, strings.Join(validThemes, ", "))
}

// validateGridSize validates the grid size preference
func (v *PreferenceValidator) validateGridSize(gridSize string) error {
	if gridSize == "" {
		return nil // Empty strings are handled at request level
	}

	validSizes := []string{"small", "medium", "large"}
	for _, valid := range validSizes {
		if gridSize == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid gridSize '%s', must be one of: %s", gridSize, strings.Join(validSizes, ", "))
}

// validateDefaultView validates the default view preference
func (v *PreferenceValidator) validateDefaultView(defaultView string) error {
	if defaultView == "" {
		return nil // Empty strings are handled at request level
	}

	validViews := []string{"grid", "list"}
	for _, valid := range validViews {
		if defaultView == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid defaultView '%s', must be one of: %s", defaultView, strings.Join(validViews, ", "))
}

// validateLanguage validates the language preference
func (v *PreferenceValidator) validateLanguage(language string) error {
	if language == "" {
		return nil // Empty strings are handled at request level
	}

	validLanguages := []string{"en", "zh-CN", "zh-TW"}
	for _, valid := range validLanguages {
		if language == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid language '%s', must be one of: %s", language, strings.Join(validLanguages, ", "))
}

// validateTimezone validates the timezone preference
func (v *PreferenceValidator) validateTimezone(timezone string) error {
	if timezone == "" {
		return nil // Empty strings are handled at request level
	}

	// Try to load the timezone to validate it
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone '%s': %w", timezone, err)
	}
	return nil
}
