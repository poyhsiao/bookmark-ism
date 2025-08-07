package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFacetedSearchParams_Validate(t *testing.T) {
	tests := []struct {
		name          string
		params        FacetedSearchParams
		expectedError string
	}{
		{
			name: "Valid parameters",
			params: FacetedSearchParams{
				Query:     "test",
				UserID:    "user123",
				FacetBy:   []string{"tags"},
				MaxFacets: 10,
				Page:      1,
				Limit:     20,
			},
		},
		{
			name: "Missing user ID",
			params: FacetedSearchParams{
				Query: "test",
				Page:  1,
				Limit: 20,
			},
			expectedError: "user_id is required",
		},
		{
			name: "Invalid page",
			params: FacetedSearchParams{
				Query:  "test",
				UserID: "user123",
				Page:   0,
				Limit:  20,
			},
			expectedError: "page must be greater than 0",
		},
		{
			name: "Invalid limit",
			params: FacetedSearchParams{
				Query:  "test",
				UserID: "user123",
				Page:   1,
				Limit:  0,
			},
			expectedError: "limit must be greater than 0",
		},
		{
			name: "Limit too high",
			params: FacetedSearchParams{
				Query:  "test",
				UserID: "user123",
				Page:   1,
				Limit:  200,
			},
			expectedError: "limit cannot exceed 100",
		},
		{
			name: "Invalid facet field",
			params: FacetedSearchParams{
				Query:   "test",
				UserID:  "user123",
				FacetBy: []string{"invalid_field"},
				Page:    1,
				Limit:   20,
			},
			expectedError: "invalid facet field: invalid_field",
		},
		{
			name: "Max facets too high",
			params: FacetedSearchParams{
				Query:     "test",
				UserID:    "user123",
				FacetBy:   []string{"tags"},
				MaxFacets: 100,
				Page:      1,
				Limit:     20,
			},
			expectedError: "max_facets cannot exceed 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSemanticSearchParams_Validate(t *testing.T) {
	tests := []struct {
		name          string
		params        SemanticSearchParams
		expectedError string
	}{
		{
			name: "Valid parameters",
			params: SemanticSearchParams{
				Query:  "machine learning",
				UserID: "user123",
				Page:   1,
				Limit:  20,
			},
		},
		{
			name: "Missing query",
			params: SemanticSearchParams{
				UserID: "user123",
				Page:   1,
				Limit:  20,
			},
			expectedError: "query is required",
		},
		{
			name: "Missing user ID",
			params: SemanticSearchParams{
				Query: "test",
				Page:  1,
				Limit: 20,
			},
			expectedError: "user_id is required",
		},
		{
			name: "Invalid page",
			params: SemanticSearchParams{
				Query:  "test",
				UserID: "user123",
				Page:   0,
				Limit:  20,
			},
			expectedError: "page must be greater than 0",
		},
		{
			name: "Invalid limit",
			params: SemanticSearchParams{
				Query:  "test",
				UserID: "user123",
				Page:   1,
				Limit:  0,
			},
			expectedError: "limit must be greater than 0",
		},
		{
			name: "Limit too high",
			params: SemanticSearchParams{
				Query:  "test",
				UserID: "user123",
				Page:   1,
				Limit:  200,
			},
			expectedError: "limit cannot exceed 100",
		},
		{
			name: "Invalid threshold - negative",
			params: SemanticSearchParams{
				Query:     "test",
				UserID:    "user123",
				Threshold: -0.1,
				Page:      1,
				Limit:     20,
			},
			// Should not error, threshold gets reset to default
		},
		{
			name: "Invalid threshold - too high",
			params: SemanticSearchParams{
				Query:     "test",
				UserID:    "user123",
				Threshold: 1.5,
				Page:      1,
				Limit:     20,
			},
			// Should not error, threshold gets reset to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSavedSearch_Validate(t *testing.T) {
	tests := []struct {
		name          string
		savedSearch   SavedSearch
		expectedError string
	}{
		{
			name: "Valid saved search",
			savedSearch: SavedSearch{
				UserID: "user123",
				Name:   "My Search",
				Query:  "golang programming",
			},
		},
		{
			name: "Missing user ID",
			savedSearch: SavedSearch{
				Name:  "My Search",
				Query: "golang programming",
			},
			expectedError: "user_id is required",
		},
		{
			name: "Missing name",
			savedSearch: SavedSearch{
				UserID: "user123",
				Query:  "golang programming",
			},
			expectedError: "name is required",
		},
		{
			name: "Missing query",
			savedSearch: SavedSearch{
				UserID: "user123",
				Name:   "My Search",
			},
			expectedError: "query is required",
		},
		{
			name: "Name too long",
			savedSearch: SavedSearch{
				UserID: "user123",
				Name:   "This is a very long name that exceeds the maximum allowed length of 100 characters for a saved search name",
				Query:  "golang programming",
			},
			expectedError: "name cannot exceed 100 characters",
		},
		{
			name: "Query too long",
			savedSearch: SavedSearch{
				UserID: "user123",
				Name:   "My Search",
				Query: "This is a very long query that exceeds the maximum allowed length of 500 characters for a saved search query. " +
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
					"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. " +
					"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
					"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			},
			expectedError: "query cannot exceed 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.savedSearch.Validate()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
