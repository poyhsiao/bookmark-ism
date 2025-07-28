package storage

import (
	"context"
	"testing"
	"time"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageClient is a mock implementation of the storage client
type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) StoreScreenshot(ctx context.Context, bookmarkID string, data []byte) (string, error) {
	args := m.Called(ctx, bookmarkID, data)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) StoreAvatar(ctx context.Context, userID string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, userID, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) StoreBackup(ctx context.Context, userID string, data []byte) (string, error) {
	args := m.Called(ctx, userID, data)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectName, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) DeleteFile(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func (m *MockStorageClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// setupTestService creates a test service with mock client
func setupTestService() (*Service, *MockStorageClient) {
	mockClient := new(MockStorageClient)
	service := &Service{
		client: mockClient,
	}
	return service, mockClient
}

// TestStoreScreenshot tests screenshot storage functionality
func TestStoreScreenshot(t *testing.T) {
	service, mockClient := setupTestService()
	ctx := context.Background()

	bookmarkID := "bookmark-123"
	data := []byte("fake screenshot data")
	expectedURL := "/storage/screenshots/bookmark-123.png"

	t.Run("Store Screenshot Successfully", func(t *testing.T) {
		mockClient.On("StoreScreenshot", ctx, bookmarkID, data).
			Return(expectedURL, nil).Once()

		url, err := service.StoreScreenshot(ctx, bookmarkID, data)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)

		mockClient.AssertExpectations(t)
	})

	t.Run("Store Screenshot Error", func(t *testing.T) {
		mockClient.On("StoreScreenshot", ctx, bookmarkID, data).
			Return("", assert.AnError).Once()

		url, err := service.StoreScreenshot(ctx, bookmarkID, data)
		assert.Error(t, err)
		assert.Empty(t, url)

		mockClient.AssertExpectations(t)
	})
}

// TestStoreAvatar tests avatar storage functionality
func TestStoreAvatar(t *testing.T) {
	service, mockClient := setupTestService()
	ctx := context.Background()

	userID := "user-456"
	data := []byte("fake avatar data")
	contentType := "image/jpeg"
	expectedURL := "/storage/avatars/user-456"

	t.Run("Store Avatar Successfully", func(t *testing.T) {
		mockClient.On("StoreAvatar", ctx, userID, data, contentType).
			Return(expectedURL, nil).Once()

		url, err := service.StoreAvatar(ctx, userID, data, contentType)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)

		mockClient.AssertExpectations(t)
	})

	t.Run("Store Avatar Error", func(t *testing.T) {
		mockClient.On("StoreAvatar", ctx, userID, data, contentType).
			Return("", assert.AnError).Once()

		url, err := service.StoreAvatar(ctx, userID, data, contentType)
		assert.Error(t, err)
		assert.Empty(t, url)

		mockClient.AssertExpectations(t)
	})
}

// TestStoreBackup tests backup storage functionality
func TestStoreBackup(t *testing.T) {
	service, mockClient := setupTestService()
	ctx := context.Background()

	userID := "user-789"
	data := []byte("fake backup data")

	t.Run("Store Backup Successfully", func(t *testing.T) {
		expectedURL := "/storage/backups/user-789/20240101-120000.json"
		mockClient.On("StoreBackup", ctx, userID, data).
			Return(expectedURL, nil).Once()

		url, err := service.StoreBackup(ctx, userID, data)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)

		mockClient.AssertExpectations(t)
	})

	t.Run("Store Backup Error", func(t *testing.T) {
		mockClient.On("StoreBackup", ctx, userID, data).
			Return("", assert.AnError).Once()

		url, err := service.StoreBackup(ctx, userID, data)
		assert.Error(t, err)
		assert.Empty(t, url)

		mockClient.AssertExpectations(t)
	})
}

// TestGetFileURLService tests file URL generation functionality in service
func TestGetFileURLService(t *testing.T) {
	service, mockClient := setupTestService()
	ctx := context.Background()

	objectName := "screenshots/bookmark-123.png"
	expiry := time.Hour
	expectedURL := "https://minio.example.com/bookmarks/screenshots/bookmark-123.png?X-Amz-Expires=3600"

	t.Run("Get File URL Successfully", func(t *testing.T) {
		mockClient.On("GetFileURL", ctx, objectName, expiry).
			Return(expectedURL, nil).Once()

		url, err := service.GetFileURL(ctx, objectName, expiry)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)

		mockClient.AssertExpectations(t)
	})

	t.Run("Get File URL Error", func(t *testing.T) {
		mockClient.On("GetFileURL", ctx, objectName, expiry).
			Return("", assert.AnError).Once()

		url, err := service.GetFileURL(ctx, objectName, expiry)
		assert.Error(t, err)
		assert.Empty(t, url)

		mockClient.AssertExpectations(t)
	})
}

// TestDeleteFileService tests file deletion functionality in service
func TestDeleteFileService(t *testing.T) {
	service, mockClient := setupTestService()
	ctx := context.Background()

	objectName := "screenshots/bookmark-123.png"

	t.Run("Delete File Successfully", func(t *testing.T) {
		mockClient.On("DeleteFile", ctx, objectName).
			Return(nil).Once()

		err := service.DeleteFile(ctx, objectName)
		assert.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Delete File Error", func(t *testing.T) {
		mockClient.On("DeleteFile", ctx, objectName).
			Return(assert.AnError).Once()

		err := service.DeleteFile(ctx, objectName)
		assert.Error(t, err)

		mockClient.AssertExpectations(t)
	})
}

// TestHealthCheckService tests health check functionality in service
func TestHealthCheckService(t *testing.T) {
	service, mockClient := setupTestService()
	ctx := context.Background()

	t.Run("Health Check Successful", func(t *testing.T) {
		mockClient.On("HealthCheck", ctx).
			Return(nil).Once()

		err := service.HealthCheck(ctx)
		assert.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Health Check Failed", func(t *testing.T) {
		mockClient.On("HealthCheck", ctx).
			Return(assert.AnError).Once()

		err := service.HealthCheck(ctx)
		assert.Error(t, err)

		mockClient.AssertExpectations(t)
	})
}

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	cfg := config.StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		UseSSL:          false,
	}

	client, err := storage.NewClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	service := NewService(client)
	assert.NotNil(t, service)
	assert.Equal(t, client, service.client)
}
