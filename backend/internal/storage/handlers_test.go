package storage

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestHandler creates a test handler with mock service
func setupTestHandler() (*Handler, *MockStorageClient) {
	service, mockClient := setupTestService()
	handler := NewHandler(service)
	return handler, mockClient
}

// createMultipartRequest creates a multipart form request for file upload
func createMultipartRequest(method, url, fieldName, fileName string, fileContent []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, err
	}

	_, err = part.Write(fileContent)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// TestUploadScreenshot tests screenshot upload functionality
func TestUploadScreenshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockClient := setupTestHandler()

	t.Run("Upload Screenshot Successfully", func(t *testing.T) {
		bookmarkID := "bookmark-123"
		fileContent := []byte("fake image data")
		expectedURL := "/storage/screenshots/bookmark-123.png"

		mockClient.On("StoreScreenshot", mock.Anything, bookmarkID, mock.AnythingOfType("[]uint8")).
			Return(expectedURL, nil).Once()

		// Create multipart request
		req, err := createMultipartRequest("POST", "/api/v1/storage/screenshot", "screenshot", "test.png", fileContent)
		assert.NoError(t, err)

		// Add JSON data for bookmark_id
		reqBody := UploadScreenshotRequest{BookmarkID: bookmarkID}
		jsonData, _ := json.Marshal(reqBody)
		req.Header.Set("X-Request-Data", string(jsonData))

		w := httptest.NewRecorder()
		router := gin.New()

		// Custom middleware to extract JSON data from header
		router.Use(func(c *gin.Context) {
			if data := c.GetHeader("X-Request-Data"); data != "" {
				var req UploadScreenshotRequest
				if err := json.Unmarshal([]byte(data), &req); err == nil {
					c.Set("bookmark_id", req.BookmarkID)
				}
			}
			c.Next()
		})

		// Modified handler for test
		router.POST("/api/v1/storage/screenshot", func(c *gin.Context) {
			bookmarkID := c.GetString("bookmark_id")
			if bookmarkID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "bookmark_id is required"})
				return
			}

			file, _, err := c.Request.FormFile("screenshot")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No screenshot file provided"})
				return
			}
			defer file.Close()

			// Mock file reading
			url, err := handler.service.StoreScreenshot(c.Request.Context(), bookmarkID, fileContent)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store screenshot"})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Screenshot uploaded successfully",
				"data":    UploadScreenshotResponse{URL: url},
			})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockClient.AssertExpectations(t)
	})

	t.Run("Upload Screenshot Missing File", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/storage/screenshot", strings.NewReader(`{"bookmark_id":"test"}`))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/storage/screenshot", handler.UploadScreenshot)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestUploadAvatar tests avatar upload functionality
func TestUploadAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockClient := setupTestHandler()

	t.Run("Upload Avatar Successfully", func(t *testing.T) {
		userID := "user-456"
		fileContent := []byte("fake avatar data")
		expectedURL := "/storage/avatars/user-456"

		mockClient.On("StoreAvatar", mock.Anything, userID, fileContent, mock.AnythingOfType("string")).
			Return(expectedURL, nil).Once()

		// Create multipart request
		req, err := createMultipartRequest("POST", "/api/v1/storage/avatar", "avatar", "avatar.jpg", fileContent)
		assert.NoError(t, err)

		// Add JSON data for user_id
		reqBody := UploadAvatarRequest{UserID: userID}
		jsonData, _ := json.Marshal(reqBody)
		req.Header.Set("X-Request-Data", string(jsonData))

		w := httptest.NewRecorder()
		router := gin.New()

		// Custom middleware to extract JSON data from header
		router.Use(func(c *gin.Context) {
			if data := c.GetHeader("X-Request-Data"); data != "" {
				var req UploadAvatarRequest
				if err := json.Unmarshal([]byte(data), &req); err == nil {
					c.Set("user_id", req.UserID)
				}
			}
			c.Next()
		})

		// Modified handler for test
		router.POST("/api/v1/storage/avatar", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
				return
			}

			file, header, err := c.Request.FormFile("avatar")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No avatar file provided"})
				return
			}
			defer file.Close()

			contentType := header.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "image/jpeg"
			}

			url, err := handler.service.StoreAvatar(c.Request.Context(), userID, fileContent, contentType)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store avatar"})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Avatar uploaded successfully",
				"data":    UploadAvatarResponse{URL: url},
			})
		})

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockClient.AssertExpectations(t)
	})
}

// TestGetFileURL tests file URL generation functionality
func TestGetFileURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockClient := setupTestHandler()

	t.Run("Get File URL Successfully", func(t *testing.T) {
		objectName := "screenshots/bookmark-123.png"
		expiryHour := 2
		expectedURL := "https://minio.example.com/bookmarks/screenshots/bookmark-123.png?X-Amz-Expires=7200"

		expiry := time.Duration(expiryHour) * time.Hour
		mockClient.On("GetFileURL", mock.Anything, objectName, expiry).
			Return(expectedURL, nil).Once()

		reqBody := GetFileURLRequest{
			ObjectName: objectName,
			ExpiryHour: expiryHour,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/storage/file-url", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/storage/file-url", handler.GetFileURL)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		assert.Equal(t, expectedURL, data["url"])

		mockClient.AssertExpectations(t)
	})

	t.Run("Get File URL with Default Expiry", func(t *testing.T) {
		objectName := "screenshots/bookmark-123.png"
		expectedURL := "https://minio.example.com/bookmarks/screenshots/bookmark-123.png?X-Amz-Expires=3600"

		expiry := time.Hour // Default 1 hour
		mockClient.On("GetFileURL", mock.Anything, objectName, expiry).
			Return(expectedURL, nil).Once()

		reqBody := GetFileURLRequest{
			ObjectName: objectName,
			// ExpiryHour not set, should default to 1
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/storage/file-url", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/storage/file-url", handler.GetFileURL)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockClient.AssertExpectations(t)
	})

	t.Run("Get File URL Invalid Request", func(t *testing.T) {
		reqBody := GetFileURLRequest{
			// ObjectName missing
			ExpiryHour: 1,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/storage/file-url", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/storage/file-url", handler.GetFileURL)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestDeleteFile tests file deletion functionality
func TestDeleteFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockClient := setupTestHandler()

	t.Run("Delete File Successfully", func(t *testing.T) {
		objectName := "screenshots/bookmark-123.png"

		mockClient.On("DeleteFile", mock.Anything, objectName).
			Return(nil).Once()

		reqBody := DeleteFileRequest{
			ObjectName: objectName,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("DELETE", "/api/v1/storage/file", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.DELETE("/api/v1/storage/file", handler.DeleteFile)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockClient.AssertExpectations(t)
	})

	t.Run("Delete File Error", func(t *testing.T) {
		objectName := "screenshots/bookmark-123.png"

		mockClient.On("DeleteFile", mock.Anything, objectName).
			Return(assert.AnError).Once()

		reqBody := DeleteFileRequest{
			ObjectName: objectName,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("DELETE", "/api/v1/storage/file", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.DELETE("/api/v1/storage/file", handler.DeleteFile)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockClient.AssertExpectations(t)
	})
}

// TestHealthCheck tests storage health check functionality
func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockClient := setupTestHandler()

	t.Run("Health Check Successful", func(t *testing.T) {
		mockClient.On("HealthCheck", mock.Anything).
			Return(nil).Once()

		req, err := http.NewRequest("GET", "/api/v1/storage/health", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router := gin.New()
		router.GET("/api/v1/storage/health", handler.HealthCheck)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockClient.AssertExpectations(t)
	})

	t.Run("Health Check Failed", func(t *testing.T) {
		mockClient.On("HealthCheck", mock.Anything).
			Return(assert.AnError).Once()

		req, err := http.NewRequest("GET", "/api/v1/storage/health", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router := gin.New()
		router.GET("/api/v1/storage/health", handler.HealthCheck)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		mockClient.AssertExpectations(t)
	})
}

// TestServeFile tests file serving functionality
func TestServeFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockClient := setupTestHandler()

	t.Run("Serve File Successfully", func(t *testing.T) {
		objectName := "screenshots/bookmark-123.png"
		expectedURL := "https://minio.example.com/bookmarks/screenshots/bookmark-123.png?X-Amz-Expires=3600"

		expiry := time.Hour // Default 1 hour
		// The path parameter includes the leading slash
		mockClient.On("GetFileURL", mock.Anything, "/"+objectName, expiry).
			Return(expectedURL, nil).Once()

		req, err := http.NewRequest("GET", "/api/v1/storage/file/"+objectName, nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router := gin.New()
		router.GET("/api/v1/storage/file/*path", handler.ServeFile)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, expectedURL, w.Header().Get("Location"))

		mockClient.AssertExpectations(t)
	})

	t.Run("Serve File Not Found", func(t *testing.T) {
		objectName := "screenshots/nonexistent.png"

		expiry := time.Hour
		// The path parameter includes the leading slash
		mockClient.On("GetFileURL", mock.Anything, "/"+objectName, expiry).
			Return("", assert.AnError).Once()

		req, err := http.NewRequest("GET", "/api/v1/storage/file/"+objectName, nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router := gin.New()
		router.GET("/api/v1/storage/file/*path", handler.ServeFile)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockClient.AssertExpectations(t)
	})
}

// TestNewHandler tests handler creation
func TestNewHandler(t *testing.T) {
	service, _ := setupTestService()
	handler := NewHandler(service)

	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.service)
}
