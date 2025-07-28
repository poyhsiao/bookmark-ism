package screenshot

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScreenshotService is a mock implementation of the screenshot service
type MockScreenshotService struct {
	mock.Mock
}

func (m *MockScreenshotService) CaptureScreenshot(ctx context.Context, bookmarkID, pageURL string, opts CaptureOptions) (*CaptureResult, error) {
	args := m.Called(ctx, bookmarkID, pageURL, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CaptureResult), args.Error(1)
}

func (m *MockScreenshotService) UpdateBookmarkScreenshot(ctx context.Context, bookmarkID, pageURL string) (*CaptureResult, error) {
	args := m.Called(ctx, bookmarkID, pageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CaptureResult), args.Error(1)
}

func (m *MockScreenshotService) GetFavicon(ctx context.Context, pageURL string) ([]byte, error) {
	args := m.Called(ctx, pageURL)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockScreenshotService) CaptureFromURL(ctx context.Context, pageURL string) ([]byte, error) {
	args := m.Called(ctx, pageURL)
	return args.Get(0).([]byte), args.Error(1)
}

// setupTestHandler creates a test handler with mock service
func setupTestHandler() (*Handler, *MockScreenshotService) {
	mockService := new(MockScreenshotService)
	handler := NewHandler(mockService)
	return handler, mockService
}

// TestCaptureScreenshot tests screenshot capture endpoint
func TestCaptureScreenshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupTestHandler()

	t.Run("Capture Screenshot Successfully", func(t *testing.T) {
		bookmarkID := "bookmark-123"
		pageURL := "https://example.com"

		expectedResult := &CaptureResult{
			URL:          "/storage/screenshots/bookmark-123.jpg",
			ThumbnailURL: "/storage/screenshots/bookmark-123_thumb.jpg",
			Width:        1200,
			Height:       800,
			Size:         12345,
			Format:       "jpeg",
		}

		mockService.On("CaptureScreenshot", mock.Anything, bookmarkID, pageURL, mock.AnythingOfType("CaptureOptions")).
			Return(expectedResult, nil).Once()

		reqBody := CaptureScreenshotRequest{
			BookmarkID: bookmarkID,
			URL:        pageURL,
			Width:      1200,
			Height:     800,
			Quality:    85,
			Format:     "jpeg",
			Thumbnail:  true,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/capture", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/capture", handler.CaptureScreenshot)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		assert.Equal(t, expectedResult.URL, data["url"])
		assert.Equal(t, expectedResult.ThumbnailURL, data["thumbnail_url"])

		mockService.AssertExpectations(t)
	})

	t.Run("Capture Screenshot Invalid Request", func(t *testing.T) {
		reqBody := CaptureScreenshotRequest{
			// Missing required fields
			URL: "https://example.com",
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/capture", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/capture", handler.CaptureScreenshot)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Capture Screenshot Service Error", func(t *testing.T) {
		bookmarkID := "bookmark-123"
		pageURL := "https://example.com"

		mockService.On("CaptureScreenshot", mock.Anything, bookmarkID, pageURL, mock.AnythingOfType("CaptureOptions")).
			Return(nil, assert.AnError).Once()

		reqBody := CaptureScreenshotRequest{
			BookmarkID: bookmarkID,
			URL:        pageURL,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/capture", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/capture", handler.CaptureScreenshot)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

// TestUpdateBookmarkScreenshot tests bookmark screenshot update endpoint
func TestUpdateBookmarkScreenshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupTestHandler()

	t.Run("Update Bookmark Screenshot Successfully", func(t *testing.T) {
		bookmarkID := "bookmark-456"
		pageURL := "https://example.com/updated"

		expectedResult := &CaptureResult{
			URL:          "/storage/screenshots/bookmark-456.jpg",
			ThumbnailURL: "/storage/screenshots/bookmark-456_thumb.jpg",
			Width:        1200,
			Height:       800,
			Size:         12345,
			Format:       "jpeg",
		}

		mockService.On("UpdateBookmarkScreenshot", mock.Anything, bookmarkID, pageURL).
			Return(expectedResult, nil).Once()

		reqBody := UpdateBookmarkScreenshotRequest{
			URL: pageURL,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("PUT", "/api/v1/screenshot/bookmark/"+bookmarkID, bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.PUT("/api/v1/screenshot/bookmark/:id", handler.UpdateBookmarkScreenshot)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockService.AssertExpectations(t)
	})

	t.Run("Update Bookmark Screenshot Missing ID", func(t *testing.T) {
		reqBody := UpdateBookmarkScreenshotRequest{
			URL: "https://example.com",
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("PUT", "/api/v1/screenshot/bookmark/", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.PUT("/api/v1/screenshot/bookmark/:id", handler.UpdateBookmarkScreenshot)

		router.ServeHTTP(w, req)

		// Gin returns 404 for missing path parameters, not 400
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestGetFavicon tests favicon retrieval endpoint
func TestGetFavicon(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupTestHandler()

	t.Run("Get Favicon Successfully", func(t *testing.T) {
		pageURL := "https://example.com"
		faviconData := []byte("fake favicon data")

		mockService.On("GetFavicon", mock.Anything, pageURL).
			Return(faviconData, nil).Once()

		reqBody := GetFaviconRequest{
			URL: pageURL,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/favicon", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/favicon", handler.GetFavicon)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/x-icon", w.Header().Get("Content-Type"))
		assert.Equal(t, faviconData, w.Body.Bytes())

		mockService.AssertExpectations(t)
	})

	t.Run("Get Favicon Not Found", func(t *testing.T) {
		pageURL := "https://example.com"

		mockService.On("GetFavicon", mock.Anything, pageURL).
			Return([]byte{}, assert.AnError).Once()

		reqBody := GetFaviconRequest{
			URL: pageURL,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/favicon", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/favicon", handler.GetFavicon)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

// TestCaptureFromURL tests direct URL screenshot capture endpoint
func TestCaptureFromURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupTestHandler()

	t.Run("Capture From URL Successfully", func(t *testing.T) {
		pageURL := "https://example.com"
		screenshotData := []byte("fake screenshot data")

		mockService.On("CaptureFromURL", mock.Anything, pageURL).
			Return(screenshotData, nil).Once()

		reqBody := CaptureFromURLRequest{
			URL: pageURL,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/url", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/url", handler.CaptureFromURL)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, screenshotData, w.Body.Bytes())

		mockService.AssertExpectations(t)
	})

	t.Run("Capture From URL Error", func(t *testing.T) {
		pageURL := "https://example.com"

		mockService.On("CaptureFromURL", mock.Anything, pageURL).
			Return([]byte{}, assert.AnError).Once()

		reqBody := CaptureFromURLRequest{
			URL: pageURL,
		}
		jsonData, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", "/api/v1/screenshot/url", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/api/v1/screenshot/url", handler.CaptureFromURL)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

// TestNewHandler tests handler creation
func TestNewHandler(t *testing.T) {
	mockService := new(MockScreenshotService)
	handler := NewHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
}
