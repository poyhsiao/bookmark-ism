package bookmark

import (
	"testing"
	"time"

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
	testUser := &database.User{
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(testUser).Error
	require.NoError(t, err)

	return db
}

func TestBookmarkService_Create(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	tests := []struct {
		name    string
		req     CreateBookmarkRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bookmark creation",
			req: CreateBookmarkRequest{
				UserID:      1,
				URL:         "https://example.com",
				Title:       "Example Website",
				Description: "A test website",
				Tags:        []string{"test", "example"},
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			req: CreateBookmarkRequest{
				UserID: 1,
				// Missing URL and Title
			},
			wantErr: true,
			errMsg:  "URL and title are required",
		},
		{
			name: "invalid URL format",
			req: CreateBookmarkRequest{
				UserID: 1,
				URL:    "not-a-valid-url",
				Title:  "Invalid URL Test",
			},
			wantErr: true,
			errMsg:  "invalid URL format",
		},
		{
			name: "non-existent user",
			req: CreateBookmarkRequest{
				UserID: 999,
				URL:    "https://example.com",
				Title:  "Test",
			},
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark, err := service.Create(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, bookmark)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bookmark)
				assert.Equal(t, tt.req.UserID, bookmark.UserID)
				assert.Equal(t, tt.req.URL, bookmark.URL)
				assert.Equal(t, tt.req.Title, bookmark.Title)
				assert.Equal(t, tt.req.Description, bookmark.Description)
				assert.NotZero(t, bookmark.ID)
				assert.NotZero(t, bookmark.CreatedAt)
			}
		})
	}
}

func TestBookmarkService_GetByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create a test bookmark
	testBookmark := &database.Bookmark{
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Test Bookmark",
		Description: "Test description",
		Status:      "active",
	}
	err := db.Create(testBookmark).Error
	require.NoError(t, err)

	tests := []struct {
		name       string
		bookmarkID uint
		userID     uint
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid bookmark retrieval",
			bookmarkID: testBookmark.ID,
			userID:     1,
			wantErr:    false,
		},
		{
			name:       "non-existent bookmark",
			bookmarkID: 999,
			userID:     1,
			wantErr:    true,
			errMsg:     "bookmark not found",
		},
		{
			name:       "unauthorized access",
			bookmarkID: testBookmark.ID,
			userID:     999,
			wantErr:    true,
			errMsg:     "bookmark not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark, err := service.GetByID(tt.bookmarkID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, bookmark)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bookmark)
				assert.Equal(t, tt.bookmarkID, bookmark.ID)
				assert.Equal(t, tt.userID, bookmark.UserID)
			}
		})
	}
}

func TestBookmarkService_Update(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create a test bookmark
	testBookmark := &database.Bookmark{
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Original Title",
		Description: "Original description",
		Status:      "active",
	}
	err := db.Create(testBookmark).Error
	require.NoError(t, err)

	tests := []struct {
		name    string
		req     UpdateBookmarkRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bookmark update",
			req: UpdateBookmarkRequest{
				ID:          testBookmark.ID,
				UserID:      1,
				Title:       "Updated Title",
				Description: "Updated description",
				Tags:        []string{"updated", "test"},
			},
			wantErr: false,
		},
		{
			name: "non-existent bookmark",
			req: UpdateBookmarkRequest{
				ID:     999,
				UserID: 1,
				Title:  "Updated Title",
			},
			wantErr: true,
			errMsg:  "bookmark not found",
		},
		{
			name: "unauthorized update",
			req: UpdateBookmarkRequest{
				ID:     testBookmark.ID,
				UserID: 999,
				Title:  "Updated Title",
			},
			wantErr: true,
			errMsg:  "bookmark not found",
		},
		{
			name: "invalid URL format",
			req: UpdateBookmarkRequest{
				ID:     testBookmark.ID,
				UserID: 1,
				URL:    "not-a-valid-url",
			},
			wantErr: true,
			errMsg:  "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark, err := service.Update(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, bookmark)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bookmark)
				assert.Equal(t, tt.req.ID, bookmark.ID)
				assert.Equal(t, tt.req.UserID, bookmark.UserID)
				if tt.req.Title != "" {
					assert.Equal(t, tt.req.Title, bookmark.Title)
				}
				if tt.req.Description != "" {
					assert.Equal(t, tt.req.Description, bookmark.Description)
				}
			}
		})
	}
}

func TestBookmarkService_Delete(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create a test bookmark
	testBookmark := &database.Bookmark{
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Test Bookmark",
		Description: "Test description",
		Status:      "active",
	}
	err := db.Create(testBookmark).Error
	require.NoError(t, err)

	tests := []struct {
		name       string
		bookmarkID uint
		userID     uint
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid bookmark deletion",
			bookmarkID: testBookmark.ID,
			userID:     1,
			wantErr:    false,
		},
		{
			name:       "non-existent bookmark",
			bookmarkID: 999,
			userID:     1,
			wantErr:    true,
			errMsg:     "bookmark not found",
		},
		{
			name:       "unauthorized deletion",
			bookmarkID: testBookmark.ID,
			userID:     999,
			wantErr:    true,
			errMsg:     "bookmark not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.bookmarkID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify bookmark is soft deleted
				var deletedBookmark database.Bookmark
				err = db.Unscoped().First(&deletedBookmark, tt.bookmarkID).Error
				assert.NoError(t, err)
				assert.NotNil(t, deletedBookmark.DeletedAt)
			}
		})
	}
}

func TestBookmarkService_List(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test bookmarks
	testBookmarks := []*database.Bookmark{
		{
			UserID:      1,
			URL:         "https://example1.com",
			Title:       "First Bookmark",
			Description: "First description",
			Status:      "active",
		},
		{
			UserID:      1,
			URL:         "https://example2.com",
			Title:       "Second Bookmark",
			Description: "Second description",
			Status:      "active",
		},
		{
			UserID:      2, // Different user
			URL:         "https://example3.com",
			Title:       "Third Bookmark",
			Description: "Third description",
			Status:      "active",
		},
	}

	for i, bookmark := range testBookmarks {
		err := db.Create(bookmark).Error
		require.NoError(t, err)

		// Update created_at to simulate different creation times
		if i == 0 {
			db.Model(bookmark).Update("created_at", time.Now().Add(-2*time.Hour))
		} else if i == 1 {
			db.Model(bookmark).Update("created_at", time.Now().Add(-1*time.Hour))
		}
	}

	tests := []struct {
		name         string
		req          ListBookmarksRequest
		expectedLen  int
		expectedURLs []string
	}{
		{
			name: "list all bookmarks for user",
			req: ListBookmarksRequest{
				UserID: 1,
				Limit:  10,
				Offset: 0,
			},
			expectedLen:  2,
			expectedURLs: []string{"https://example2.com", "https://example1.com"}, // Ordered by created_at DESC
		},
		{
			name: "list with limit",
			req: ListBookmarksRequest{
				UserID: 1,
				Limit:  1,
				Offset: 0,
			},
			expectedLen:  1,
			expectedURLs: []string{"https://example2.com"},
		},
		{
			name: "list with offset",
			req: ListBookmarksRequest{
				UserID: 1,
				Limit:  10,
				Offset: 1,
			},
			expectedLen:  1,
			expectedURLs: []string{"https://example1.com"},
		},
		{
			name: "search by title",
			req: ListBookmarksRequest{
				UserID: 1,
				Search: "First",
				Limit:  10,
				Offset: 0,
			},
			expectedLen:  1,
			expectedURLs: []string{"https://example1.com"},
		},
		{
			name: "no results for different user",
			req: ListBookmarksRequest{
				UserID: 999,
				Limit:  10,
				Offset: 0,
			},
			expectedLen:  0,
			expectedURLs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarks, total, err := service.List(tt.req)

			assert.NoError(t, err)
			assert.Len(t, bookmarks, tt.expectedLen)

			if tt.expectedLen > 0 {
				// Total should be the total count after filtering but before pagination
				expectedTotal := int64(len(tt.expectedURLs))
				if tt.name == "list all bookmarks for user" || tt.name == "list with limit" || tt.name == "list with offset" {
					expectedTotal = 2 // We have 2 bookmarks for user 1
				}
				assert.Equal(t, expectedTotal, total)
				for i, expectedURL := range tt.expectedURLs {
					if i < len(bookmarks) {
						assert.Equal(t, expectedURL, bookmarks[i].URL)
					}
				}
			}
		})
	}
}
