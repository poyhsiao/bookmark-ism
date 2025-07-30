package import_export

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"bookmark-sync-service/backend/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
}

func TestService_ImportBookmarksFromChrome(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)

	// Create test user
	user := &database.User{
		BaseModel:   database.BaseModel{ID: userID},
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Chrome bookmarks JSON format
	chromeBookmarks := `{
		"checksum": "test-checksum",
		"roots": {
			"bookmark_bar": {
				"children": [
					{
						"date_added": "13285932710000000",
						"guid": "test-guid-1",
						"id": "1",
						"name": "Google",
						"type": "url",
						"url": "https://www.google.com"
					},
					{
						"children": [
							{
								"date_added": "13285932720000000",
								"guid": "test-guid-2",
								"id": "2",
								"name": "GitHub",
								"type": "url",
								"url": "https://github.com"
							}
						],
						"date_added": "13285932700000000",
						"date_modified": "13285932720000000",
						"guid": "test-folder-guid",
						"id": "3",
						"name": "Development",
						"type": "folder"
					}
				],
				"date_added": "13285932700000000",
				"date_modified": "13285932720000000",
				"guid": "bookmark_bar_guid",
				"id": "0",
				"name": "Bookmarks bar",
				"type": "folder"
			}
		},
		"version": 1
	}`

	// Test import
	result, err := service.ImportBookmarksFromChrome(ctx, userID, strings.NewReader(chromeBookmarks))
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ImportedBookmarksCount)
	assert.Equal(t, 1, result.ImportedCollectionsCount)
	assert.Equal(t, 0, result.DuplicatesSkipped)
	assert.Len(t, result.Errors, 0)

	// Verify bookmarks were created
	var bookmarks []database.Bookmark
	err = db.Where("user_id = ?", userID).Find(&bookmarks).Error
	require.NoError(t, err)
	assert.Len(t, bookmarks, 2)

	// Verify collections were created
	var collections []database.Collection
	err = db.Where("user_id = ?", userID).Find(&collections).Error
	require.NoError(t, err)
	assert.Len(t, collections, 1)
	assert.Equal(t, "Development", collections[0].Name)
}

func TestService_ImportBookmarksFromFirefox(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)

	// Create test user
	user := &database.User{
		BaseModel:   database.BaseModel{ID: userID},
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Firefox bookmarks HTML format
	firefoxBookmarks := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks Menu</H1>
<DL><p>
    <DT><H3 ADD_DATE="1640995200" LAST_MODIFIED="1640995300">Development</H3>
    <DL><p>
        <DT><A HREF="https://github.com" ADD_DATE="1640995200">GitHub</A>
        <DT><A HREF="https://stackoverflow.com" ADD_DATE="1640995250">Stack Overflow</A>
    </DL><p>
    <DT><A HREF="https://www.google.com" ADD_DATE="1640995100">Google</A>
</DL><p>`

	// Test import
	result, err := service.ImportBookmarksFromFirefox(ctx, userID, strings.NewReader(firefoxBookmarks))
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.ImportedBookmarksCount)
	assert.Equal(t, 1, result.ImportedCollectionsCount)
	assert.Equal(t, 0, result.DuplicatesSkipped)
	assert.Len(t, result.Errors, 0)

	// Verify bookmarks were created
	var bookmarks []database.Bookmark
	err = db.Where("user_id = ?", userID).Find(&bookmarks).Error
	require.NoError(t, err)
	assert.Len(t, bookmarks, 3)

	// Verify collections were created
	var collections []database.Collection
	err = db.Where("user_id = ?", userID).Find(&collections).Error
	require.NoError(t, err)
	assert.Len(t, collections, 1)
	assert.Equal(t, "Development", collections[0].Name)
}

func TestService_ImportBookmarksFromSafari(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)

	// Create test user
	user := &database.User{
		BaseModel:   database.BaseModel{ID: userID},
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Safari bookmarks plist format (simplified)
	safariBookmarks := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Children</key>
	<array>
		<dict>
			<key>Title</key>
			<string>BookmarksBar</string>
			<key>Children</key>
			<array>
				<dict>
					<key>URLString</key>
					<string>https://www.google.com</string>
					<key>URIDictionary</key>
					<dict>
						<key>title</key>
						<string>Google</string>
					</dict>
				</dict>
				<dict>
					<key>Title</key>
					<string>Development</string>
					<key>Children</key>
					<array>
						<dict>
							<key>URLString</key>
							<string>https://github.com</string>
							<key>URIDictionary</key>
							<dict>
								<key>title</key>
								<string>GitHub</string>
							</dict>
						</dict>
					</array>
				</dict>
			</array>
		</dict>
	</array>
</dict>
</plist>`

	// Test import
	result, err := service.ImportBookmarksFromSafari(ctx, userID, strings.NewReader(safariBookmarks))
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ImportedBookmarksCount)
	assert.Equal(t, 1, result.ImportedCollectionsCount)
	assert.Equal(t, 0, result.DuplicatesSkipped)
	assert.Len(t, result.Errors, 0)
}

func TestService_ExportBookmarksToJSON(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)

	// Create test user
	user := &database.User{
		BaseModel:   database.BaseModel{ID: userID},
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Create test collection
	collection := &database.Collection{
		BaseModel:   database.BaseModel{ID: 1},
		UserID:      userID,
		Name:        "Test Collection",
		Description: "A test collection",
		Visibility:  "private",
	}
	err = db.Create(collection).Error
	require.NoError(t, err)

	// Create test bookmarks
	bookmarks := []*database.Bookmark{
		{
			BaseModel:   database.BaseModel{ID: 1},
			UserID:      userID,
			URL:         "https://www.google.com",
			Title:       "Google",
			Description: "Search engine",
			Tags:        `["search", "google"]`,
		},
		{
			BaseModel:   database.BaseModel{ID: 2},
			UserID:      userID,
			URL:         "https://github.com",
			Title:       "GitHub",
			Description: "Code repository",
			Tags:        `["development", "git"]`,
		},
	}

	for _, bookmark := range bookmarks {
		err = db.Create(bookmark).Error
		require.NoError(t, err)

		// Associate with collection
		err = db.Model(collection).Association("Bookmarks").Append(bookmark)
		require.NoError(t, err)
	}

	// Test export
	var result strings.Builder
	err = service.ExportBookmarksToJSON(ctx, userID, &result)
	require.NoError(t, err)

	// Verify JSON structure
	var exportData map[string]interface{}
	err = json.Unmarshal([]byte(result.String()), &exportData)
	require.NoError(t, err)

	assert.Contains(t, exportData, "bookmarks")
	assert.Contains(t, exportData, "collections")
	assert.Contains(t, exportData, "exported_at")
	assert.Contains(t, exportData, "version")

	bookmarksData := exportData["bookmarks"].([]interface{})
	assert.Len(t, bookmarksData, 2)

	collectionsData := exportData["collections"].([]interface{})
	assert.Len(t, collectionsData, 1)
}

func TestService_ExportBookmarksToHTML(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)

	// Create test user
	user := &database.User{
		BaseModel:   database.BaseModel{ID: userID},
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Create test bookmarks
	bookmarks := []*database.Bookmark{
		{
			BaseModel:   database.BaseModel{ID: 1},
			UserID:      userID,
			URL:         "https://www.google.com",
			Title:       "Google",
			Description: "Search engine",
		},
		{
			BaseModel:   database.BaseModel{ID: 2},
			UserID:      userID,
			URL:         "https://github.com",
			Title:       "GitHub",
			Description: "Code repository",
		},
	}

	for _, bookmark := range bookmarks {
		err = db.Create(bookmark).Error
		require.NoError(t, err)
	}

	// Test export
	var result strings.Builder
	err = service.ExportBookmarksToHTML(ctx, userID, &result)
	require.NoError(t, err)

	htmlContent := result.String()
	assert.Contains(t, htmlContent, "<!DOCTYPE NETSCAPE-Bookmark-file-1>")
	assert.Contains(t, htmlContent, "Google")
	assert.Contains(t, htmlContent, "GitHub")
	assert.Contains(t, htmlContent, "https://www.google.com")
	assert.Contains(t, htmlContent, "https://github.com")
}

func TestService_DetectDuplicates(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)

	// Create test user
	user := &database.User{
		BaseModel:   database.BaseModel{ID: userID},
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		SupabaseID:  "test-supabase-id",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Create existing bookmark
	existingBookmark := &database.Bookmark{
		BaseModel: database.BaseModel{ID: 1},
		UserID:    userID,
		URL:       "https://www.google.com",
		Title:     "Google",
	}
	err = db.Create(existingBookmark).Error
	require.NoError(t, err)

	// Test duplicate detection
	isDuplicate, err := service.DetectDuplicate(ctx, userID, "https://www.google.com")
	require.NoError(t, err)
	assert.True(t, isDuplicate)

	// Test non-duplicate
	isDuplicate, err = service.DetectDuplicate(ctx, userID, "https://github.com")
	require.NoError(t, err)
	assert.False(t, isDuplicate)
}

func TestService_GetImportProgress(t *testing.T) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)
	defer database.CleanupTestDB(db)

	service := NewService(db)
	ctx := context.Background()
	userID := uint(1)
	jobID := "test-job-123"

	// Test getting progress for non-existent job
	progress, err := service.GetImportProgress(ctx, userID, jobID)
	require.NoError(t, err)
	assert.Nil(t, progress)

	// Note: In this simplified implementation, we don't actually store progress
	// In a real implementation, you would use Redis or a database table
	// For now, we just test that the function doesn't crash
}

func TestImportResult_Validate(t *testing.T) {
	tests := []struct {
		name   string
		result ImportResult
		valid  bool
	}{
		{
			name: "valid result",
			result: ImportResult{
				ImportedBookmarksCount:   10,
				ImportedCollectionsCount: 2,
				DuplicatesSkipped:        1,
				Errors:                   []string{},
			},
			valid: true,
		},
		{
			name: "negative bookmarks count",
			result: ImportResult{
				ImportedBookmarksCount:   -1,
				ImportedCollectionsCount: 2,
				DuplicatesSkipped:        1,
				Errors:                   []string{},
			},
			valid: false,
		},
		{
			name: "negative collections count",
			result: ImportResult{
				ImportedBookmarksCount:   10,
				ImportedCollectionsCount: -1,
				DuplicatesSkipped:        1,
				Errors:                   []string{},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
