package storage

import (
	"context"
	"io"
	"net/url"
	"testing"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMinioClient is a mock implementation of the MinIO client
// MockMinioClient 是 MinIO 客戶端的模擬實現
type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (m *MockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, size, opts)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

func (m *MockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*minio.Object), args.Error(1)
}

func (m *MockMinioClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

func (m *MockMinioClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams map[string]string) (*url.URL, error) {
	args := m.Called(ctx, bucketName, objectName, expiry, reqParams)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*url.URL), args.Error(1)
}

func (m *MockMinioClient) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	args := m.Called(ctx, bucketName, opts)
	return args.Get(0).(<-chan minio.ObjectInfo)
}

// setupTestClient creates a test client with mock MinIO client
// setupTestClient 創建帶有模擬 MinIO 客戶端的測試客戶端
func setupTestClient() (*Client, *MockMinioClient) {
	mockClient := new(MockMinioClient)

	cfg := config.StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		UseSSL:          false,
	}

	client := &Client{
		client:     mockClient,
		config:     &cfg,
		bucketName: cfg.BucketName,
	}

	return client, mockClient
}

// TestHealthCheck tests the health check functionality
// TestHealthCheck 測試健康檢查功能
func TestHealthCheck(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	t.Run("Healthy", func(t *testing.T) {
		mockClient.On("BucketExists", ctx, "test-bucket").Return(true, nil).Once()

		err := client.HealthCheck(ctx)
		assert.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Unhealthy", func(t *testing.T) {
		mockClient.On("BucketExists", ctx, "test-bucket").Return(false, assert.AnError).Once()

		err := client.HealthCheck(ctx)
		assert.Error(t, err)

		mockClient.AssertExpectations(t)
	})
}

// TestEnsureBucketExists tests the bucket creation functionality
// TestEnsureBucketExists 測試存儲桶創建功能
func TestEnsureBucketExists(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	t.Run("Bucket Already Exists", func(t *testing.T) {
		mockClient.On("BucketExists", ctx, "test-bucket").Return(true, nil).Once()

		err := client.EnsureBucketExists(ctx)
		assert.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Create New Bucket", func(t *testing.T) {
		mockClient.On("BucketExists", ctx, "test-bucket").Return(false, nil).Once()
		mockClient.On("MakeBucket", ctx, "test-bucket", mock.Anything).Return(nil).Once()

		err := client.EnsureBucketExists(ctx)
		assert.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Error Checking Bucket", func(t *testing.T) {
		mockClient.On("BucketExists", ctx, "test-bucket").Return(false, assert.AnError).Once()

		err := client.EnsureBucketExists(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check if bucket exists")

		mockClient.AssertExpectations(t)
	})

	t.Run("Error Creating Bucket", func(t *testing.T) {
		mockClient.On("BucketExists", ctx, "test-bucket").Return(false, nil).Once()
		mockClient.On("MakeBucket", ctx, "test-bucket", mock.Anything).Return(assert.AnError).Once()

		err := client.EnsureBucketExists(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create bucket")

		mockClient.AssertExpectations(t)
	})
}

// TestUploadFile tests the file upload functionality
// TestUploadFile 測試文件上傳功能
func TestUploadFile(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	objectName := "test/object.txt"
	data := []byte("test data")
	contentType := "text/plain"

	t.Run("Upload Successfully", func(t *testing.T) {
		uploadInfo := minio.UploadInfo{
			Bucket:       "test-bucket",
			Key:          objectName,
			ETag:         "test-etag",
			Size:         int64(len(data)),
			LastModified: time.Now(),
		}

		mockClient.On("PutObject", ctx, "test-bucket", objectName, mock.Anything, int64(len(data)), mock.Anything).
			Return(uploadInfo, nil).Once()

		url, err := client.UploadFile(ctx, objectName, data, contentType)
		assert.NoError(t, err)
		assert.Equal(t, "/storage/test/object.txt", url)

		mockClient.AssertExpectations(t)
	})

	t.Run("Upload Error", func(t *testing.T) {
		uploadInfo := minio.UploadInfo{}

		mockClient.On("PutObject", ctx, "test-bucket", objectName, mock.Anything, int64(len(data)), mock.Anything).
			Return(uploadInfo, assert.AnError).Once()

		url, err := client.UploadFile(ctx, objectName, data, contentType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upload file")
		assert.Empty(t, url)

		mockClient.AssertExpectations(t)
	})
}

// TestStoreScreenshot tests the screenshot storage functionality
// TestStoreScreenshot 測試截圖存儲功能
func TestStoreScreenshot(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	bookmarkID := "123"
	data := []byte("fake screenshot data")
	expectedObjectName := "screenshots/123.png"

	t.Run("Store Screenshot Successfully", func(t *testing.T) {
		uploadInfo := minio.UploadInfo{
			Bucket:       "test-bucket",
			Key:          expectedObjectName,
			ETag:         "test-etag",
			Size:         int64(len(data)),
			LastModified: time.Now(),
		}

		mockClient.On("PutObject", ctx, "test-bucket", expectedObjectName, mock.Anything, int64(len(data)), mock.Anything).
			Return(uploadInfo, nil).Once()

		url, err := client.StoreScreenshot(ctx, bookmarkID, data)
		assert.NoError(t, err)
		assert.Equal(t, "/storage/screenshots/123.png", url)

		mockClient.AssertExpectations(t)
	})
}

// TestStoreAvatar tests the avatar storage functionality
// TestStoreAvatar 測試頭像存儲功能
func TestStoreAvatar(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	userID := "456"
	data := []byte("fake avatar data")
	contentType := "image/jpeg"
	expectedObjectName := "avatars/456"

	t.Run("Store Avatar Successfully", func(t *testing.T) {
		uploadInfo := minio.UploadInfo{
			Bucket:       "test-bucket",
			Key:          expectedObjectName,
			ETag:         "test-etag",
			Size:         int64(len(data)),
			LastModified: time.Now(),
		}

		mockClient.On("PutObject", ctx, "test-bucket", expectedObjectName, mock.Anything, int64(len(data)), mock.Anything).
			Return(uploadInfo, nil).Once()

		url, err := client.StoreAvatar(ctx, userID, data, contentType)
		assert.NoError(t, err)
		assert.Equal(t, "/storage/avatars/456", url)

		mockClient.AssertExpectations(t)
	})
}

// TestStoreBackup tests the backup storage functionality
// TestStoreBackup 測試備份存儲功能
func TestStoreBackup(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	userID := "789"
	data := []byte("fake backup data")

	t.Run("Store Backup Successfully", func(t *testing.T) {
		uploadInfo := minio.UploadInfo{
			Bucket:       "test-bucket",
			Key:          mock.Anything,
			ETag:         "test-etag",
			Size:         int64(len(data)),
			LastModified: time.Now(),
		}

		mockClient.On("PutObject", ctx, "test-bucket", mock.AnythingOfType("string"), mock.Anything, int64(len(data)), mock.Anything).
			Return(uploadInfo, nil).Once()

		url, err := client.StoreBackup(ctx, userID, data)
		assert.NoError(t, err)
		assert.Contains(t, url, "/storage/backups/789/")
		assert.Contains(t, url, ".json")

		mockClient.AssertExpectations(t)
	})
}

// TestStoreBucketFile tests bucket-specific file storage
func TestStoreBucketFile(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	bucketType := "documents"
	fileName := "test.pdf"
	data := []byte("fake pdf data")
	contentType := "application/pdf"
	expectedObjectName := "documents/test.pdf"

	t.Run("Store Bucket File Successfully", func(t *testing.T) {
		uploadInfo := minio.UploadInfo{
			Bucket:       "test-bucket",
			Key:          expectedObjectName,
			ETag:         "test-etag",
			Size:         int64(len(data)),
			LastModified: time.Now(),
		}

		mockClient.On("PutObject", ctx, "test-bucket", expectedObjectName, mock.Anything, int64(len(data)), mock.Anything).
			Return(uploadInfo, nil).Once()

		url, err := client.StoreBucketFile(ctx, bucketType, fileName, data, contentType)
		assert.NoError(t, err)
		assert.Equal(t, "/storage/documents/test.pdf", url)

		mockClient.AssertExpectations(t)
	})
}

// TestGetBucketFiles tests listing files in a bucket directory
func TestGetBucketFiles(t *testing.T) {
	client, mockClient := setupTestClient()
	ctx := context.Background()

	bucketType := "screenshots"
	expectedFiles := []string{"screenshots/file1.jpg", "screenshots/file2.png"}

	t.Run("Get Bucket Files Successfully", func(t *testing.T) {
		objectCh := make(chan minio.ObjectInfo, 2)
		go func() {
			defer close(objectCh)
			objectCh <- minio.ObjectInfo{Key: "screenshots/file1.jpg"}
			objectCh <- minio.ObjectInfo{Key: "screenshots/file2.png"}
		}()

		mockClient.On("ListObjects", ctx, "test-bucket", mock.MatchedBy(func(opts minio.ListObjectsOptions) bool {
			return opts.Prefix == "screenshots/"
		})).Return((<-chan minio.ObjectInfo)(objectCh)).Once()

		files, err := client.GetBucketFiles(ctx, bucketType)
		assert.NoError(t, err)
		assert.Equal(t, expectedFiles, files)

		mockClient.AssertExpectations(t)
	})
}
