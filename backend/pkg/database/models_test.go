package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
// 創建用於測試的內存 SQLite 資料庫
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	// 執行遷移
	err = AutoMigrate(db)
	require.NoError(t, err)

	return db
}

// TestUserModel tests the User model functionality
// 測試 User 模型功能
func TestUserModel(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Create User", func(t *testing.T) {
		user := User{
			Email:       "test@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
			SupabaseID:  "test-supabase-id",
			Preferences: `{"theme": "dark"}`,
		}

		err := db.Create(&user).Error
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("Unique Constraints", func(t *testing.T) {
		// Create first user
		// 創建第一個用戶
		user1 := User{
			Email:      "unique@example.com",
			Username:   "uniqueuser",
			SupabaseID: "unique-supabase-id",
		}
		err := db.Create(&user1).Error
		assert.NoError(t, err)

		// Try to create user with same email
		// 嘗試創建具有相同電子郵件的用戶
		user2 := User{
			Email:      "unique@example.com",
			Username:   "differentuser",
			SupabaseID: "different-supabase-id",
		}
		err = db.Create(&user2).Error
		assert.Error(t, err)

		// Try to create user with same username
		// 嘗試創建具有相同用戶名的用戶
		user3 := User{
			Email:      "different@example.com",
			Username:   "uniqueuser",
			SupabaseID: "another-supabase-id",
		}
		err = db.Create(&user3).Error
		assert.Error(t, err)
	})

	t.Run("User Relationships", func(t *testing.T) {
		// Create user
		// 創建用戶
		user := User{
			Email:      "relations@example.com",
			Username:   "relationsuser",
			SupabaseID: "relations-supabase-id",
		}
		err := db.Create(&user).Error
		require.NoError(t, err)

		// Create bookmark
		// 創建書籤
		bookmark := Bookmark{
			UserID:      user.ID,
			URL:         "https://example.com",
			Title:       "Example Site",
			Description: "An example website",
			Status:      "active",
		}
		err = db.Create(&bookmark).Error
		require.NoError(t, err)

		// Create collection
		// 創建收藏夾
		collection := Collection{
			UserID:     user.ID,
			Name:       "Test Collection",
			Visibility: "private",
		}
		err = db.Create(&collection).Error
		require.NoError(t, err)

		// Load user with relationships
		// 載入用戶及其關聯
		var loadedUser User
		err = db.Preload("Bookmarks").Preload("Collections").First(&loadedUser, user.ID).Error
		require.NoError(t, err)

		assert.Len(t, loadedUser.Bookmarks, 1)
		assert.Len(t, loadedUser.Collections, 1)
		assert.Equal(t, "Example Site", loadedUser.Bookmarks[0].Title)
		assert.Equal(t, "Test Collection", loadedUser.Collections[0].Name)
	})
}

// TestBookmarkModel tests the Bookmark model functionality
// 測試 Bookmark 模型功能
func TestBookmarkModel(t *testing.T) {
	db := setupTestDB(t)

	// Create user first
	// 首先創建用戶
	user := User{
		Email:      "bookmark@example.com",
		Username:   "bookmarkuser",
		SupabaseID: "bookmark-supabase-id",
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	t.Run("Create Bookmark", func(t *testing.T) {
		bookmark := Bookmark{
			UserID:      user.ID,
			URL:         "https://golang.org",
			Title:       "The Go Programming Language",
			Description: "Go is an open source programming language",
			Tags:        `["programming", "golang", "language"]`,
			Status:      "active",
		}

		err := db.Create(&bookmark).Error
		assert.NoError(t, err)
		assert.NotZero(t, bookmark.ID)
		assert.Equal(t, "active", bookmark.Status)
	})

	t.Run("Bookmark with Collections", func(t *testing.T) {
		// Create collection
		// 創建收藏夾
		collection := Collection{
			UserID:     user.ID,
			Name:       "Programming Resources",
			Visibility: "private",
			ShareLink:  "programming-resources-test",
		}
		err := db.Create(&collection).Error
		require.NoError(t, err)

		// Create bookmark
		// 創建書籤
		bookmark := Bookmark{
			UserID: user.ID,
			URL:    "https://github.com",
			Title:  "GitHub",
			Status: "active",
		}
		err = db.Create(&bookmark).Error
		require.NoError(t, err)

		// Associate bookmark with collection
		// 將書籤與收藏夾關聯
		err = db.Model(&collection).Association("Bookmarks").Append(&bookmark)
		require.NoError(t, err)

		// Load collection with bookmarks
		// 載入收藏夾及其書籤
		var loadedCollection Collection
		err = db.Preload("Bookmarks").First(&loadedCollection, collection.ID).Error
		require.NoError(t, err)

		assert.Len(t, loadedCollection.Bookmarks, 1)
		assert.Equal(t, "GitHub", loadedCollection.Bookmarks[0].Title)
	})
}

// TestCollectionModel tests the Collection model functionality
// 測試 Collection 模型功能
func TestCollectionModel(t *testing.T) {
	db := setupTestDB(t)

	// Create user first
	// 首先創建用戶
	user := User{
		Email:      "collection@example.com",
		Username:   "collectionuser",
		SupabaseID: "collection-supabase-id",
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	t.Run("Create Collection", func(t *testing.T) {
		collection := Collection{
			UserID:      user.ID,
			Name:        "My Collection",
			Description: "A test collection",
			Color:       "#FF0000",
			Icon:        "folder",
			Visibility:  "private",
			ShareLink:   "my-collection-test",
		}

		err := db.Create(&collection).Error
		assert.NoError(t, err)
		assert.NotZero(t, collection.ID)
		assert.Equal(t, "private", collection.Visibility)
	})

	t.Run("Hierarchical Collections", func(t *testing.T) {
		// Create parent collection
		// 創建父收藏夾
		parent := Collection{
			UserID:     user.ID,
			Name:       "Parent Collection",
			Visibility: "private",
			ShareLink:  "parent-collection-test",
		}
		err := db.Create(&parent).Error
		require.NoError(t, err)

		// Create child collection
		// 創建子收藏夾
		child := Collection{
			UserID:     user.ID,
			Name:       "Child Collection",
			ParentID:   &parent.ID,
			Visibility: "private",
			ShareLink:  "child-collection-test",
		}
		err = db.Create(&child).Error
		require.NoError(t, err)

		// Load parent with children
		// 載入父收藏夾及其子項
		var loadedParent Collection
		err = db.Preload("Children").First(&loadedParent, parent.ID).Error
		require.NoError(t, err)

		assert.Len(t, loadedParent.Children, 1)
		assert.Equal(t, "Child Collection", loadedParent.Children[0].Name)

		// Load child with parent
		// 載入子收藏夾及其父項
		var loadedChild Collection
		err = db.Preload("Parent").First(&loadedChild, child.ID).Error
		require.NoError(t, err)

		assert.NotNil(t, loadedChild.Parent)
		assert.Equal(t, "Parent Collection", loadedChild.Parent.Name)
	})
}

// TestSeedTestData tests the seed data functionality
// 測試種子數據功能
func TestSeedTestData(t *testing.T) {
	db := setupTestDB(t)

	err := SeedTestData(db)
	assert.NoError(t, err)

	// Check that users were created
	// 檢查用戶是否已創建
	var userCount int64
	err = db.Model(&User{}).Count(&userCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), userCount)

	// Check that collections were created
	// 檢查收藏夾是否已創建
	var collectionCount int64
	err = db.Model(&Collection{}).Count(&collectionCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), collectionCount)

	// Check that bookmarks were created
	// 檢查書籤是否已創建
	var bookmarkCount int64
	err = db.Model(&Bookmark{}).Count(&bookmarkCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(4), bookmarkCount)

	// Try to seed again (should fail)
	// 嘗試再次播種（應該失敗）
	err = SeedTestData(db)
	assert.Error(t, err)
}

// TestAutoMigrate tests the database migration functionality
// 測試資料庫遷移功能
func TestAutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = AutoMigrate(db)
	assert.NoError(t, err)

	// Check that tables were created
	// 檢查表是否已創建
	tables := []string{"users", "bookmarks", "collections", "comments", "sync_events", "follows"}
	for _, table := range tables {
		assert.True(t, db.Migrator().HasTable(table), "Table %s should exist", table)
	}
}

// TestRollback tests the database rollback functionality
// 測試資料庫回滾功能
func TestRollback(t *testing.T) {
	db := setupTestDB(t)

	// Tables should exist after migration
	// 遷移後表應該存在
	assert.True(t, db.Migrator().HasTable("users"))
	assert.True(t, db.Migrator().HasTable("bookmarks"))

	// Rollback
	// 回滾
	err := Rollback(db)
	assert.NoError(t, err)

	// Tables should not exist after rollback
	// 回滾後表不應該存在
	assert.False(t, db.Migrator().HasTable("users"))
	assert.False(t, db.Migrator().HasTable("bookmarks"))
}
