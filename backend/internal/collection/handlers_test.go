package collection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	// Create test user
	user := &database.User{
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	require.NoError(t, db.Create(user).Error)

	router := gin.New()

	// Mock auth middleware that sets user_id to 1
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	// Setup collection routes
	service := NewService(db)
	handler := NewHandler(service)

	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	return router, db
}

func TestHandler_CreateCollection(t *testing.T) {
	router, _ := setupTestRouter(t)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid collection creation",
			requestBody: CreateCollectionRequest{
				Name:        "Test Collection",
				Description: "A test collection",
				Color:       "#FF5733",
				Icon:        "folder",
				Visibility:  "private",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			requestBody: CreateCollectionRequest{
				Description: "Missing name",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
		{
			name: "invalid visibility",
			requestBody: CreateCollectionRequest{
				Name:       "Test Collection",
				Visibility: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/collections", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError != "" {
				assert.False(t, response["success"].(bool))
				errorObj := response["error"].(map[string]interface{})
				assert.Contains(t, errorObj["message"], tt.expectedError)
			} else {
				assert.True(t, response["success"].(bool))
				assert.Equal(t, "Collection created successfully", response["message"])
				assert.NotNil(t, response["data"])
			}
		})
	}
}

func TestHandler_ListCollections(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collections
	collections := []CreateCollectionRequest{
		{Name: "Collection A", Visibility: "private"},
		{Name: "Collection B", Visibility: "public"},
		{Name: "Collection C", Visibility: "private"},
	}

	for _, req := range collections {
		_, err := service.Create(1, req)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all collections",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "list with pagination",
			queryParams:    "?page=1&limit=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "filter by visibility",
			queryParams:    "?visibility=public",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "search collections",
			queryParams:    "?search=Collection%20A",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "invalid page parameter",
			queryParams:    "?page=0",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collections"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				data := response["data"].(map[string]interface{})
				collections := data["collections"].([]interface{})
				assert.Len(t, collections, tt.expectedCount)
			}
		})
	}
}

func TestHandler_GetCollection(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collection
	req := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	created, err := service.Create(1, req)
	require.NoError(t, err)

	tests := []struct {
		name           string
		collectionID   string
		expectedStatus int
	}{
		{
			name:           "get existing collection",
			collectionID:   fmt.Sprintf("%d", created.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent collection",
			collectionID:   "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid collection ID",
			collectionID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collections/"+tt.collectionID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Collection retrieved successfully", response["message"])
				assert.NotNil(t, response["data"])
			}
		})
	}
}

func TestHandler_UpdateCollection(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collection
	req := CreateCollectionRequest{
		Name:       "Original Name",
		Visibility: "private",
	}
	created, err := service.Create(1, req)
	require.NoError(t, err)

	tests := []struct {
		name           string
		collectionID   string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:         "valid update",
			collectionID: fmt.Sprintf("%d", created.ID),
			requestBody: UpdateCollectionRequest{
				Name:        func() *string { s := "Updated Name"; return &s }(),
				Description: func() *string { s := "Updated Description"; return &s }(),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "empty name should fail",
			collectionID: fmt.Sprintf("%d", created.ID),
			requestBody: UpdateCollectionRequest{
				Name: func() *string { s := ""; return &s }(),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent collection",
			collectionID:   "999",
			requestBody:    UpdateCollectionRequest{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid collection ID",
			collectionID:   "invalid",
			requestBody:    UpdateCollectionRequest{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/collections/"+tt.collectionID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Collection updated successfully", response["message"])
				assert.NotNil(t, response["data"])
			}
		})
	}
}

func TestHandler_DeleteCollection(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collection
	req := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	created, err := service.Create(1, req)
	require.NoError(t, err)

	tests := []struct {
		name           string
		collectionID   string
		expectedStatus int
	}{
		{
			name:           "valid deletion",
			collectionID:   fmt.Sprintf("%d", created.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent collection",
			collectionID:   "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid collection ID",
			collectionID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/collections/"+tt.collectionID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Collection deleted successfully", response["message"])
			}
		})
	}
}

func TestHandler_AddBookmarkToCollection(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collection
	collectionReq := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	collection, err := service.Create(1, collectionReq)
	require.NoError(t, err)

	// Create test bookmark
	bookmark := &database.Bookmark{
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Test Bookmark",
		Description: "A test bookmark",
		Status:      "active",
	}
	require.NoError(t, db.Create(bookmark).Error)

	tests := []struct {
		name           string
		collectionID   string
		bookmarkID     string
		expectedStatus int
	}{
		{
			name:           "valid bookmark addition",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			bookmarkID:     fmt.Sprintf("%d", bookmark.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent collection",
			collectionID:   "999",
			bookmarkID:     fmt.Sprintf("%d", bookmark.ID),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-existent bookmark",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			bookmarkID:     "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid collection ID",
			collectionID:   "invalid",
			bookmarkID:     fmt.Sprintf("%d", bookmark.ID),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid bookmark ID",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			bookmarkID:     "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/collections/%s/bookmarks/%s", tt.collectionID, tt.bookmarkID)
			req := httptest.NewRequest(http.MethodPost, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Bookmark added to collection successfully", response["message"])
			}
		})
	}
}

func TestHandler_RemoveBookmarkFromCollection(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collection
	collectionReq := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	collection, err := service.Create(1, collectionReq)
	require.NoError(t, err)

	// Create test bookmark
	bookmark := &database.Bookmark{
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Test Bookmark",
		Description: "A test bookmark",
		Status:      "active",
	}
	require.NoError(t, db.Create(bookmark).Error)

	// Add bookmark to collection first
	err = service.AddBookmark(1, collection.ID, bookmark.ID)
	require.NoError(t, err)

	tests := []struct {
		name           string
		collectionID   string
		bookmarkID     string
		expectedStatus int
	}{
		{
			name:           "valid bookmark removal",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			bookmarkID:     fmt.Sprintf("%d", bookmark.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent collection",
			collectionID:   "999",
			bookmarkID:     fmt.Sprintf("%d", bookmark.ID),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid collection ID",
			collectionID:   "invalid",
			bookmarkID:     fmt.Sprintf("%d", bookmark.ID),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid bookmark ID",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			bookmarkID:     "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/collections/%s/bookmarks/%s", tt.collectionID, tt.bookmarkID)
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Bookmark removed from collection successfully", response["message"])
			}
		})
	}
}

func TestHandler_GetCollectionBookmarks(t *testing.T) {
	router, db := setupTestRouter(t)
	service := NewService(db)

	// Create test collection
	collectionReq := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	collection, err := service.Create(1, collectionReq)
	require.NoError(t, err)

	// Create test bookmarks
	bookmarks := []*database.Bookmark{
		{
			UserID:      1,
			URL:         "https://example1.com",
			Title:       "Bookmark 1",
			Description: "First bookmark",
			Status:      "active",
		},
		{
			UserID:      1,
			URL:         "https://example2.com",
			Title:       "Bookmark 2",
			Description: "Second bookmark",
			Status:      "active",
		},
	}

	for _, bookmark := range bookmarks {
		require.NoError(t, db.Create(bookmark).Error)
		err = service.AddBookmark(1, collection.ID, bookmark.ID)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		collectionID   string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get all bookmarks",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "get with pagination",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			queryParams:    "?page=1&limit=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "search bookmarks",
			collectionID:   fmt.Sprintf("%d", collection.ID),
			queryParams:    "?search=First",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "non-existent collection",
			collectionID:   "999",
			queryParams:    "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid collection ID",
			collectionID:   "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/collections/%s/bookmarks%s", tt.collectionID, tt.queryParams)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				data := response["data"].(map[string]interface{})
				bookmarks := data["bookmarks"].([]interface{})
				assert.Len(t, bookmarks, tt.expectedCount)
			}
		})
	}
}
