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
	// Use test name to create unique users
	testName := t.Name()
	user := &database.User{
		Email:       fmt.Sprintf("test-%s@example.com", testName),
		Username:    fmt.Sprintf("testuser-%s", testName),
		DisplayName: "Test User",
		SupabaseID:  fmt.Sprintf("test-supabase-id-%s", testName),
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

// TestUpdatePreferencesService tests the UpdatePreferences service functionality with validation
// TestUpdatePreferencesService 測試 UpdatePreferences 服務功能及其驗證
func TestUpdatePreferencesService(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	t.Run("Update Preferences Successfully", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Valid preferences update
		// 有效的偏好設置更新
		req := &UpdatePreferencesRequest{
			Theme:       "dark",
			GridSize:    "large",
			DefaultView: "list",
			Language:    "zh-CN",
			Timezone:    "Asia/Shanghai",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, profile)

		// Verify preferences were updated
		// 驗證偏好設置已更新
		assert.Equal(t, "dark", profile.Preferences.Theme)
		assert.Equal(t, "large", profile.Preferences.GridSize)
		assert.Equal(t, "list", profile.Preferences.DefaultView)
		assert.Equal(t, "zh-CN", profile.Preferences.Language)
		assert.Equal(t, "Asia/Shanghai", profile.Preferences.Timezone)
	})

	t.Run("Update Preferences with Invalid Theme", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Invalid theme value
		// 無效的主題值
		req := &UpdatePreferencesRequest{
			Theme: "invalid_theme",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "invalid theme")
		assert.Contains(t, err.Error(), "must be one of: light, dark, auto")
	})

	t.Run("Update Preferences with Invalid Grid Size", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Invalid grid size value
		// 無效的網格大小值
		req := &UpdatePreferencesRequest{
			GridSize: "extra_large",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "invalid gridSize")
		assert.Contains(t, err.Error(), "must be one of: small, medium, large")
	})

	t.Run("Update Preferences with Invalid Default View", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Invalid default view value
		// 無效的默認視圖值
		req := &UpdatePreferencesRequest{
			DefaultView: "card_view",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "invalid defaultView")
		assert.Contains(t, err.Error(), "must be one of: grid, list")
	})

	t.Run("Update Preferences with Invalid Language", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Invalid language value
		// 無效的語言值
		req := &UpdatePreferencesRequest{
			Language: "fr",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "invalid language")
		assert.Contains(t, err.Error(), "must be one of: en, zh-CN, zh-TW")
	})

	t.Run("Update Preferences with Invalid Timezone Format", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Invalid timezone format
		// 無效的時區格式
		req := &UpdatePreferencesRequest{
			Timezone: "Invalid/Timezone",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "invalid timezone")
	})

	t.Run("Update Preferences with Multiple Invalid Values", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Multiple invalid values
		// 多個無效值
		req := &UpdatePreferencesRequest{
			Theme:       "neon",
			GridSize:    "tiny",
			DefaultView: "carousel",
			Language:    "klingon",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		assert.Error(t, err)
		assert.Nil(t, profile)

		// Should contain multiple validation errors
		// 應包含多個驗證錯誤
		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "validation failed")
		assert.Contains(t, errorMsg, "theme")
		assert.Contains(t, errorMsg, "gridSize")
		assert.Contains(t, errorMsg, "defaultView")
		assert.Contains(t, errorMsg, "language")
	})

	t.Run("Update Preferences for Non-existent User", func(t *testing.T) {
		// Valid preferences but non-existent user
		// 有效的偏好設置但用戶不存在
		req := &UpdatePreferencesRequest{
			Theme: "dark",
		}

		profile, err := service.UpdatePreferences(ctx, 999, req)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("Update Preferences with Empty Values Should Not Change Existing", func(t *testing.T) {
		// Create test user with existing preferences
		// 創建帶有現有偏好設置的測試用戶
		user := createTestUser(t, db)

		// Set initial preferences
		// 設置初始偏好設置
		initialReq := &UpdatePreferencesRequest{
			Theme:    "dark",
			GridSize: "large",
			Language: "zh-CN",
		}

		_, err := service.UpdatePreferences(ctx, user.ID, initialReq)
		require.NoError(t, err)

		// Update with empty values (should not change existing)
		// 使用空值更新（不應更改現有設置）
		emptyReq := &UpdatePreferencesRequest{}

		profile, err := service.UpdatePreferences(ctx, user.ID, emptyReq)
		require.NoError(t, err)

		// Verify existing preferences are preserved
		// 驗證現有偏好設置被保留
		assert.Equal(t, "dark", profile.Preferences.Theme)
		assert.Equal(t, "large", profile.Preferences.GridSize)
		assert.Equal(t, "zh-CN", profile.Preferences.Language)
	})

	t.Run("Update Preferences with Partial Valid Values", func(t *testing.T) {
		// Create test user
		// 創建測試用戶
		user := createTestUser(t, db)

		// Update only some preferences
		// 僅更新部分偏好設置
		req := &UpdatePreferencesRequest{
			Theme:    "auto",
			Language: "zh-TW",
		}

		profile, err := service.UpdatePreferences(ctx, user.ID, req)
		require.NoError(t, err)

		// Verify updated values
		// 驗證更新的值
		assert.Equal(t, "auto", profile.Preferences.Theme)
		assert.Equal(t, "zh-TW", profile.Preferences.Language)

		// Verify default values for non-updated fields
		// 驗證未更新字段的默認值
		assert.Equal(t, "medium", profile.Preferences.GridSize)
		assert.Equal(t, "grid", profile.Preferences.DefaultView)
	})
}
