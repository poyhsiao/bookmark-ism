package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	return db
}

func TestCollectionService_Create(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	tests := []struct {
		name        string
		userID      uint
		req         CreateCollectionRequest
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid collection creation",
			userID: 1,
			req: CreateCollectionRequest{
				Name:        "Test Collection",
				Description: "A test collection",
				Color:       "#FF5733",
				Icon:        "folder",
				Visibility:  "private",
			},
			wantErr: false,
		},
		{
			name:   "collection with parent",
			userID: 1,
			req: CreateCollectionRequest{
				Name:        "Sub Collection",
				Description: "A sub collection",
				ParentID:    func() *uint { id := uint(1); return &id }(),
				Visibility:  "private",
			},
			wantErr: false,
		},
		{
			name:   "empty name should fail",
			userID: 1,
			req: CreateCollectionRequest{
				Name:       "",
				Visibility: "private",
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:   "invalid visibility should fail",
			userID: 1,
			req: CreateCollectionRequest{
				Name:       "Test Collection",
				Visibility: "invalid",
			},
			wantErr:     true,
			errContains: "invalid visibility",
		},
		{
			name:   "non-existent parent should fail",
			userID: 1,
			req: CreateCollectionRequest{
				Name:       "Test Collection",
				ParentID:   func() *uint { id := uint(999); return &id }(),
				Visibility: "private",
			},
			wantErr:     true,
			errContains: "parent collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parent collection if needed for the second test
			if tt.name == "collection with parent" {
				parentReq := CreateCollectionRequest{
					Name:       "Parent Collection",
					Visibility: "private",
				}
				_, err := service.Create(tt.userID, parentReq)
				require.NoError(t, err)
			}

			collection, err := service.Create(tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, collection)
			assert.Equal(t, tt.req.Name, collection.Name)
			assert.Equal(t, tt.req.Description, collection.Description)
			assert.Equal(t, tt.req.Color, collection.Color)
			assert.Equal(t, tt.req.Icon, collection.Icon)
			assert.Equal(t, tt.req.Visibility, collection.Visibility)
			assert.Equal(t, tt.userID, collection.UserID)
			assert.NotEmpty(t, collection.ShareLink)
			assert.NotZero(t, collection.ID)
			assert.NotZero(t, collection.CreatedAt)
			assert.NotZero(t, collection.UpdatedAt)

			if tt.req.ParentID != nil {
				assert.Equal(t, *tt.req.ParentID, *collection.ParentID)
			}
		})
	}
}

func TestCollectionService_GetByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test collection
	req := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	created, err := service.Create(1, req)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uint
		id          uint
		wantErr     bool
		errContains string
	}{
		{
			name:    "existing collection",
			userID:  1,
			id:      created.ID,
			wantErr: false,
		},
		{
			name:        "non-existent collection",
			userID:      1,
			id:          999,
			wantErr:     true,
			errContains: "collection not found",
		},
		{
			name:        "unauthorized access",
			userID:      999,
			id:          created.ID,
			wantErr:     true,
			errContains: "collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection, err := service.GetByID(tt.userID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, collection)
			assert.Equal(t, tt.id, collection.ID)
			assert.Equal(t, tt.userID, collection.UserID)
		})
	}
}

func TestCollectionService_List(t *testing.T) {
	db := setupTestDB(t)
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
		name    string
		userID  uint
		params  ListCollectionsParams
		wantLen int
		wantErr bool
	}{
		{
			name:   "list all collections",
			userID: 1,
			params: ListCollectionsParams{
				Page:  1,
				Limit: 10,
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:   "list with pagination",
			userID: 1,
			params: ListCollectionsParams{
				Page:  1,
				Limit: 2,
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:   "filter by visibility",
			userID: 1,
			params: ListCollectionsParams{
				Page:       1,
				Limit:      10,
				Visibility: "public",
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:   "search by name",
			userID: 1,
			params: ListCollectionsParams{
				Page:   1,
				Limit:  10,
				Search: "Collection A",
			},
			wantLen: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.List(tt.userID, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result.Collections, tt.wantLen)
			assert.Equal(t, tt.params.Page, result.Page)
			assert.Equal(t, tt.params.Limit, result.Limit)
			assert.GreaterOrEqual(t, result.Total, int64(tt.wantLen))
		})
	}
}

func TestCollectionService_Update(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test collection
	req := CreateCollectionRequest{
		Name:        "Original Name",
		Description: "Original Description",
		Visibility:  "private",
	}
	created, err := service.Create(1, req)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uint
		id          uint
		req         UpdateCollectionRequest
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid update",
			userID: 1,
			id:     created.ID,
			req: UpdateCollectionRequest{
				Name:        func() *string { s := "Updated Name"; return &s }(),
				Description: func() *string { s := "Updated Description"; return &s }(),
				Color:       func() *string { s := "#FF0000"; return &s }(),
			},
			wantErr: false,
		},
		{
			name:   "partial update",
			userID: 1,
			id:     created.ID,
			req: UpdateCollectionRequest{
				Name: func() *string { s := "Partially Updated"; return &s }(),
			},
			wantErr: false,
		},
		{
			name:   "empty name should fail",
			userID: 1,
			id:     created.ID,
			req: UpdateCollectionRequest{
				Name: func() *string { s := ""; return &s }(),
			},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "non-existent collection",
			userID:      1,
			id:          999,
			req:         UpdateCollectionRequest{},
			wantErr:     true,
			errContains: "collection not found",
		},
		{
			name:        "unauthorized update",
			userID:      999,
			id:          created.ID,
			req:         UpdateCollectionRequest{},
			wantErr:     true,
			errContains: "collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection, err := service.Update(tt.userID, tt.id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, collection)
			assert.Equal(t, tt.id, collection.ID)
			assert.Equal(t, tt.userID, collection.UserID)

			if tt.req.Name != nil {
				assert.Equal(t, *tt.req.Name, collection.Name)
			}
			if tt.req.Description != nil {
				assert.Equal(t, *tt.req.Description, collection.Description)
			}
			if tt.req.Color != nil {
				assert.Equal(t, *tt.req.Color, collection.Color)
			}
		})
	}
}

func TestCollectionService_Delete(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test collection
	req := CreateCollectionRequest{
		Name:       "Test Collection",
		Visibility: "private",
	}
	created, err := service.Create(1, req)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uint
		id          uint
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid deletion",
			userID:  1,
			id:      created.ID,
			wantErr: false,
		},
		{
			name:        "non-existent collection",
			userID:      1,
			id:          999,
			wantErr:     true,
			errContains: "collection not found",
		},
		{
			name:        "unauthorized deletion",
			userID:      999,
			id:          created.ID,
			wantErr:     true,
			errContains: "collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.userID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Verify collection is soft deleted
			var collection database.Collection
			err = db.Unscoped().First(&collection, tt.id).Error
			require.NoError(t, err)
			assert.NotNil(t, collection.DeletedAt)
		})
	}
}

func TestCollectionService_AddBookmark(t *testing.T) {
	db := setupTestDB(t)
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
		name         string
		userID       uint
		collectionID uint
		bookmarkID   uint
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid bookmark addition",
			userID:       1,
			collectionID: collection.ID,
			bookmarkID:   bookmark.ID,
			wantErr:      false,
		},
		{
			name:         "duplicate addition should be idempotent",
			userID:       1,
			collectionID: collection.ID,
			bookmarkID:   bookmark.ID,
			wantErr:      false,
		},
		{
			name:         "non-existent collection",
			userID:       1,
			collectionID: 999,
			bookmarkID:   bookmark.ID,
			wantErr:      true,
			errContains:  "collection not found",
		},
		{
			name:         "non-existent bookmark",
			userID:       1,
			collectionID: collection.ID,
			bookmarkID:   999,
			wantErr:      true,
			errContains:  "bookmark not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AddBookmark(tt.userID, tt.collectionID, tt.bookmarkID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Verify bookmark is associated with collection
			var updatedCollection database.Collection
			err = db.Preload("Bookmarks").First(&updatedCollection, tt.collectionID).Error
			require.NoError(t, err)

			found := false
			for _, b := range updatedCollection.Bookmarks {
				if b.ID == tt.bookmarkID {
					found = true
					break
				}
			}
			assert.True(t, found, "bookmark should be associated with collection")
		})
	}
}

func TestCollectionService_RemoveBookmark(t *testing.T) {
	db := setupTestDB(t)
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
		name         string
		userID       uint
		collectionID uint
		bookmarkID   uint
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid bookmark removal",
			userID:       1,
			collectionID: collection.ID,
			bookmarkID:   bookmark.ID,
			wantErr:      false,
		},
		{
			name:         "removing non-existent association should not error",
			userID:       1,
			collectionID: collection.ID,
			bookmarkID:   bookmark.ID,
			wantErr:      false,
		},
		{
			name:         "non-existent collection",
			userID:       1,
			collectionID: 999,
			bookmarkID:   bookmark.ID,
			wantErr:      true,
			errContains:  "collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RemoveBookmark(tt.userID, tt.collectionID, tt.bookmarkID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Verify bookmark is no longer associated with collection
			var updatedCollection database.Collection
			err = db.Preload("Bookmarks").First(&updatedCollection, tt.collectionID).Error
			require.NoError(t, err)

			found := false
			for _, b := range updatedCollection.Bookmarks {
				if b.ID == tt.bookmarkID {
					found = true
					break
				}
			}
			assert.False(t, found, "bookmark should not be associated with collection")
		})
	}
}

func TestCollectionService_GetBookmarks(t *testing.T) {
	db := setupTestDB(t)
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
		name         string
		userID       uint
		collectionID uint
		params       GetCollectionBookmarksParams
		wantLen      int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "get all bookmarks",
			userID:       1,
			collectionID: collection.ID,
			params: GetCollectionBookmarksParams{
				Page:  1,
				Limit: 10,
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:         "get with pagination",
			userID:       1,
			collectionID: collection.ID,
			params: GetCollectionBookmarksParams{
				Page:  1,
				Limit: 1,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:         "search bookmarks",
			userID:       1,
			collectionID: collection.ID,
			params: GetCollectionBookmarksParams{
				Page:   1,
				Limit:  10,
				Search: "First",
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:         "non-existent collection",
			userID:       1,
			collectionID: 999,
			params: GetCollectionBookmarksParams{
				Page:  1,
				Limit: 10,
			},
			wantErr:     true,
			errContains: "collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetBookmarks(tt.userID, tt.collectionID, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Len(t, result.Bookmarks, tt.wantLen)
			assert.Equal(t, tt.params.Page, result.Page)
			assert.Equal(t, tt.params.Limit, result.Limit)
			assert.GreaterOrEqual(t, result.Total, int64(tt.wantLen))
		})
	}
}
