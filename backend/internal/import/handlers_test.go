package import_export

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *Service) {
	gin.SetMode(gin.TestMode)

	db, _ := database.SetupTestDB()
	service := NewService(db)
	handlers := NewHandlers(service)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock user ID for testing
		c.Set("user_id", uint(1))
		c.Next()
	})

	// Register routes
	v1 := router.Group("/api/v1")
	handlers.RegisterRoutes(v1)

	return router, service
}

func TestHandlers_ImportFromChrome(t *testing.T) {
	router, _ := setupTestRouter()

	// Create test Chrome bookmarks JSON
	chromeBookmarks := `{
		"checksum": "test-checksum",
		"roots": {
			"bookmark_bar": {
				"children": [
					{
						"date_added": "13285932710000000",
						"guid": "test-guid-1",
						"id": "1",
						"name": "Google",
						"type": "url",
						"url": "https://www.google.com"
					}
				],
				"date_added": "13285932700000000",
				"guid": "bookmark_bar_guid",
				"id": "0",
				"name": "Bookmarks bar",
				"type": "folder"
			}
		},
		"version": 1
	}`

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "bookmarks.json")
	require.NoError(t, err)

	_, err = part.Write([]byte(chromeBookmarks))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import-export/import/chrome", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}
func TestHandlers_ExportToJSON(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/import-export/export/json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
}

func TestHandlers_DetectDuplicates(t *testing.T) {
	router, _ := setupTestRouter()

	requestBody := map[string]interface{}{
		"urls": []string{
			"https://www.google.com",
			"https://github.com",
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import-export/detect-duplicates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		function func(string) bool
		expected bool
	}{
		{"HTML file with .html extension", "bookmarks.html", isHTMLFile, true},
		{"HTML file with .htm extension", "bookmarks.htm", isHTMLFile, true},
		{"Non-HTML file", "bookmarks.json", isHTMLFile, false},
		{"Plist file", "bookmarks.plist", isPlistFile, true},
		{"Non-plist file", "bookmarks.html", isPlistFile, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}
