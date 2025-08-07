package bookmark

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/utils"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	// Create test user
	testUser := &database.User{
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(testUser).Error
	require.NoError(t, err)

	// Setup router
	router := gin.New()
	service := NewService(db)
	handlers := NewHandlers(service)

	// Mock auth middleware that sets user ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	// Setup routes
	v1 := router.Group("/api/v1")
	handlers.RegisterRoutes(v1)

	return router, db
}

func TestCreateBookmark(t *testing.T) {
	router, _ := setupTestRouter(t)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid bookmark creation",
			requestBody: CreateBookmarkRequest{
				URL:         "https://example.com",
				Title:       "Example Website",
				Description: "A test website",
				Tags:        []string{"test", "example"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			requestBody: CreateBookmarkRequest{
				Description: "Missing URL and title",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "URL and title are required",
		},
		{
			name: "invalid URL format",
			requestBody: CreateBookmarkRequest{
				URL:   "not-a-valid-url",
				Title: "Invalid URL Test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/v1/bookmarks", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response utils.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			} else if tt.expectedStatus == http.StatusCreated {
				var response utils.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "Bookmark created successfully", response.Message)
				assert.NotNil(t, response.Data)
			}
		})
	}
}

func TestGetBookmark(t *testing.T) {
	router, db := setupTestRouter(t)

	// Create a test bookmark
	testBookmark := &database.Bookmark{
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Test Bookmark",
		Description: "Test description",
		Status:      "active",
	}
	err := db.Create(testBookmark).Error
	require.NoError(t, err)

	tests := []struct {
		name           string
		bookmarkID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid bookmark retrieval",
			bookmarkID:     "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent bookmark",
			bookmarkID:     "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "bookmark not found",
		},
		{
			name:           "invalid bookmark ID",
			bookmarkID:     "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/v1/bookmarks/"+tt.bookmarkID, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response utils.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response utils.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			}
		})
	}
}

func TestListBookmarks(t *testing.T) {
	router, db := setupTestRouter(t)

	// Create test bookmarks
	testBookmarks := []*database.Bookmark{
		{
			UserID:      1,
			URL:         "https://example1.com",
			Title:       "First Bookmark",
			Description: "First description",
			Status:      "active",
		},
		{
			UserID:      1,
			URL:         "https://example2.com",
			Title:       "Second Bookmark",
			Description: "Second description",
			Status:      "active",
		},
	}

	for _, bookmark := range testBookmarks {
		err := db.Create(bookmark).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all bookmarks",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "list with limit",
			queryParams:    "?limit=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "search bookmarks",
			queryParams:    "?search=First",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/v1/bookmarks"+tt.queryParams, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response utils.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)

				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)

				bookmarks, ok := data["bookmarks"].([]interface{})
				require.True(t, ok)
				assert.Len(t, bookmarks, tt.expectedCount)
			}
		})
	}
}
