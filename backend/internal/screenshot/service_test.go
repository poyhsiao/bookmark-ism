package screenshot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService is a mock implementation of the storage service
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) StoreScreenshot(ctx context.Context, bookmarkID string, data []byte) (string, error) {
	args := m.Called(ctx, bookmarkID, data)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) StoreAvatar(ctx context.Context, userID string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, userID, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) StoreBackup(ctx context.Context, userID string, data []byte) (string, error) {
	args := m.Called(ctx, userID, data)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectName, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) DeleteFile(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func (m *MockStorageService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// setupTestService creates a test service with mock storage
func setupTestService() (*Service, *MockStorageService) {
	mockStorage := new(MockStorageService)
	service := NewService(mockStorage)
	return service, mockStorage
}

// TestCaptureScreenshotService tests screenshot capture functionality in service
func TestCaptureScreenshotService(t *testing.T) {
	service, mockStorage := setupTestService()
	ctx := context.Background()

	bookmarkID := "bookmark-123"
	pageURL := "https://example.com"
	expectedURL := "/storage/screenshots/bookmark-123.png"
	expectedThumbnailURL := "/storage/screenshots/bookmark-123_thumb.png"

	opts := CaptureOptions{
		Width:     1200,
		Height:    800,
		Quality:   85,
		Format:    "jpeg",
		Thumbnail: true,
	}

	t.Run("Capture Screenshot Successfully", func(t *testing.T) {
		// Mock storage calls
		mockStorage.On("StoreScreenshot", ctx, bookmarkID, mock.AnythingOfType("[]uint8")).
			Return(expectedURL, nil).Once()
		mockStorage.On("StoreScreenshot", ctx, bookmarkID+"_thumb", mock.AnythingOfType("[]uint8")).
			Return(expectedThumbnailURL, nil).Once()

		result, err := service.CaptureScreenshot(ctx, bookmarkID, pageURL, opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedURL, result.URL)
		assert.Equal(t, expectedThumbnailURL, result.ThumbnailURL)
		assert.Equal(t, opts.Width, result.Width)
		assert.Equal(t, opts.Height, result.Height)
		assert.Equal(t, opts.Format, result.Format)
		assert.Greater(t, result.Size, int64(0))

		mockStorage.AssertExpectations(t)
	})

	t.Run("Capture Screenshot Without Thumbnail", func(t *testing.T) {
		optsNoThumb := opts
		optsNoThumb.Thumbnail = false

		mockStorage.On("StoreScreenshot", ctx, bookmarkID, mock.AnythingOfType("[]uint8")).
			Return(expectedURL, nil).Once()

		result, err := service.CaptureScreenshot(ctx, bookmarkID, pageURL, optsNoThumb)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedURL, result.URL)
		assert.Empty(t, result.ThumbnailURL)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Capture Screenshot Invalid URL", func(t *testing.T) {
		invalidURL := "://invalid-url"

		result, err := service.CaptureScreenshot(ctx, bookmarkID, invalidURL, opts)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid URL")
	})

	t.Run("Capture Screenshot Storage Error", func(t *testing.T) {
		mockStorage.On("StoreScreenshot", ctx, bookmarkID, mock.AnythingOfType("[]uint8")).
			Return("", assert.AnError).Once()

		result, err := service.CaptureScreenshot(ctx, bookmarkID, pageURL, opts)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to store screenshot")

		mockStorage.AssertExpectations(t)
	})
}

// TestCaptureFromURLService tests URL-based screenshot capture in service
func TestCaptureFromURLService(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	pageURL := "https://example.com"

	t.Run("Capture From URL Successfully", func(t *testing.T) {
		data, err := service.CaptureFromURL(ctx, pageURL)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Greater(t, len(data), 0)
	})
}

// TestGeneratePlaceholderScreenshot tests placeholder generation
func TestGeneratePlaceholderScreenshot(t *testing.T) {
	service, _ := setupTestService()

	pageURL := "https://example.com"
	opts := CaptureOptions{
		Width:  800,
		Height: 600,
		Format: "png",
	}

	t.Run("Generate Placeholder Successfully", func(t *testing.T) {
		data, err := service.generatePlaceholderScreenshot(pageURL, opts)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Greater(t, len(data), 0)

		// Check PNG signature
		assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, data[:4])
	})
}

// TestGenerateThumbnail tests thumbnail generation
func TestGenerateThumbnail(t *testing.T) {
	service, _ := setupTestService()

	screenshotData := []byte("fake screenshot data")
	width := 300
	height := 200

	t.Run("Generate Thumbnail Successfully", func(t *testing.T) {
		thumbnailData, err := service.generateThumbnail(screenshotData, width, height)
		assert.NoError(t, err)
		assert.NotNil(t, thumbnailData)
		assert.Greater(t, len(thumbnailData), 0)
	})
}

// TestUpdateBookmarkScreenshotService tests bookmark screenshot update in service
func TestUpdateBookmarkScreenshotService(t *testing.T) {
	service, mockStorage := setupTestService()
	ctx := context.Background()

	bookmarkID := "bookmark-456"
	pageURL := "https://example.com/page"
	expectedURL := "/storage/screenshots/bookmark-456.png"
	expectedThumbnailURL := "/storage/screenshots/bookmark-456_thumb.png"

	t.Run("Update Bookmark Screenshot Successfully", func(t *testing.T) {
		mockStorage.On("StoreScreenshot", ctx, bookmarkID, mock.AnythingOfType("[]uint8")).
			Return(expectedURL, nil).Once()
		mockStorage.On("StoreScreenshot", ctx, bookmarkID+"_thumb", mock.AnythingOfType("[]uint8")).
			Return(expectedThumbnailURL, nil).Once()

		result, err := service.UpdateBookmarkScreenshot(ctx, bookmarkID, pageURL)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedURL, result.URL)
		assert.Equal(t, expectedThumbnailURL, result.ThumbnailURL)
		assert.Equal(t, 1200, result.Width)
		assert.Equal(t, 800, result.Height)
		assert.Equal(t, "jpeg", result.Format)

		mockStorage.AssertExpectations(t)
	})
}

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	mockStorage := new(MockStorageService)
	service := NewService(mockStorage)

	assert.NotNil(t, service)
	assert.Equal(t, mockStorage, service.storageService)
	assert.NotNil(t, service.httpClient)
	assert.Equal(t, 30*time.Second, service.httpClient.Timeout)
}
