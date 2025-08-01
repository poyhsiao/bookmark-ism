package user

import (
	"context"
	"fmt"
	"testing"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockStorageClient is a mock implementation of the storage client
// MockStorageClient 是存儲客戶端的模擬實現
type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) UploadFile(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, objectName, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageClient) EnsureBucketExists(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// setupTestService creates a test service with in-memory database
// setupTestService 創建帶有內存資料庫的測試服務
func setupTestService(t *testing.T) (*Service, *gorm.DB, *MockStorageClient) {
	// Setup in-memory database
	// 設置內存資料庫
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	// 執行遷移
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	// Create mock storage client
	// 創建模擬存儲客戶端
	mockStorage := &MockStorageClient{}

	// Create logger
	// 創建日誌記錄器
	logger := zap.NewNop()

	// Create service
	// 創建服務
	service := NewService(db, mockStorage, logger)

	return service, db, mockStorage
}

// createTestUser creates a test user in the database
// createTestUser 在資料庫中創建測試用戶
func createTestUser(t *testing.T, db *gorm.DB) *database.User {
	user := &database.User{
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
		Preferences: `{"theme": "light", "gridSize": "medium"}`,
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// TestUploadAvatar tests the UploadAvatar functionality
// TestUploadAvatar 測試 UploadAvatar 功能
func TestUploadAvatar(t *testing.T) {
	service, db, mockStorage := setupTestService(t)
	ctx := context.Background()

	t.Run("Upload Avatar Successfully", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Mock storage upload
		// 模擬存儲上傳
		mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), "image/png").
			Return("https://storage.example.com/avatar.png", nil)

		// Upload avatar
		// 上傳頭像
		imageData := []byte("fake image data")
		profile, err := service.UploadAvatar(ctx, user.ID, imageData, "image/png")
		require.NoError(t, err)

		assert.Equal(t, "https://storage.example.com/avatar.png", profile.Avatar)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Upload Avatar for Non-existent User", func(t *testing.T) {
		imageData := []byte("fake image data")
		profile, err := service.UploadAvatar(ctx, 999, imageData, "image/png")
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")
	})
}

// TestExportUserData tests the ExportUserData functionality
// TestExportUserData 測試 ExportUserData 功能
func TestExportUserData(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	t.Run("Export User Data Successfully", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Create test bookmark
		// 創建測試書籤
		bookmark := &database.Bookmark{
			UserID: user.ID,
			URL:    "https://example.com",
			Title:  "Example",
			Status: "active",
		}
		err := db.Create(bookmark).Error
		require.NoError(t, err)

		// Create test collection
		// 創建測試收藏夾
		collection := &database.Collection{
			UserID:     user.ID,
			Name:       "Test Collection",
			Visibility: "private",
			ShareLink:  "test-collection-export",
		}
		err = db.Create(collection).Error
		require.NoError(t, err)

		// Export data
		// 導出數據
		exportData, err := service.ExportUserData(ctx, user.ID)
		require.NoError(t, err)

		assert.Contains(t, exportData, "user_profile")
		assert.Contains(t, exportData, "bookmarks")
		assert.Contains(t, exportData, "collections")
		assert.Contains(t, exportData, "statistics")
		assert.Contains(t, exportData, "export_date")

		// Check user profile data
		// 檢查用戶個人資料數據
		userProfile := exportData["user_profile"].(map[string]interface{})
		assert.Equal(t, user.Email, userProfile["email"])
		assert.Equal(t, user.Username, userProfile["username"])

		// Check bookmarks data
		// 檢查書籤數據
		bookmarks := exportData["bookmarks"].([]database.Bookmark)
		assert.Len(t, bookmarks, 1)
		assert.Equal(t, "Example", bookmarks[0].Title)

		// Check collections data
		// 檢查收藏夾數據
		collections := exportData["collections"].([]database.Collection)
		assert.Len(t, collections, 1)
		assert.Equal(t, "Test Collection", collections[0].Name)
	})
}

// TestDeleteUser tests the DeleteUser functionality
// TestDeleteUser 測試 DeleteUser 功能
func TestDeleteUser(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	t.Run("Delete User Successfully", func(t *testing.T) {
		// Create test user with related data
		// 創建帶有相關數據的測試用戶
		user := createTestUser(t, db)

		// Create related data
		// 創建相關數據
		bookmark := &database.Bookmark{
			UserID: user.ID,
			URL:    "https://example.com",
			Title:  "Example",
			Status: "active",
		}
		err := db.Create(bookmark).Error
		require.NoError(t, err)

		collection := &database.Collection{
			UserID:     user.ID,
			Name:       "Test Collection",
			Visibility: "private",
			ShareLink:  "test-collection-delete",
		}
		err = db.Create(collection).Error
		require.NoError(t, err)

		// Delete user
		// 刪除用戶
		err = service.DeleteUser(ctx, user.ID)
		require.NoError(t, err)

		// Verify user is deleted
		// 驗證用戶已刪除
		var deletedUser database.User
		err = db.First(&deletedUser, user.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify related data is deleted
		// 驗證相關數據已刪除
		var bookmarkCount int64
		err = db.Model(&database.Bookmark{}).Where("user_id = ?", user.ID).Count(&bookmarkCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), bookmarkCount)

		var collectionCount int64
		err = db.Model(&database.Collection{}).Where("user_id = ?", user.ID).Count(&collectionCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), collectionCount)
	})
}

// TestGetUserStats tests the getUserStats functionality
// TestGetUserStats 測試 getUserStats 功能
func TestGetUserStats(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	t.Run("Get User Stats", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Create test data
		// 創建測試數據
		for i := 0; i < 3; i++ {
			bookmark := &database.Bookmark{
				UserID: user.ID,
				URL:    "https://example.com",
				Title:  "Example",
				Status: "active",
			}
			err := db.Create(bookmark).Error
			require.NoError(t, err)
		}

		for i := 0; i < 2; i++ {
			collection := &database.Collection{
				UserID:     user.ID,
				Name:       fmt.Sprintf("Test Collection %d", i+1),
				Visibility: "private",
				ShareLink:  fmt.Sprintf("test-share-link-%d-%d", user.ID, i),
			}
			err := db.Create(collection).Error
			require.NoError(t, err)
		}

		// Get stats
		// 獲取統計信息
		stats, err := service.getUserStats(ctx, user.ID)
		require.NoError(t, err)

		assert.Equal(t, 3, stats.BookmarkCount)
		assert.Equal(t, 2, stats.CollectionCount)
		assert.Equal(t, 0, stats.StorageUsed) // Default value
	})
}
