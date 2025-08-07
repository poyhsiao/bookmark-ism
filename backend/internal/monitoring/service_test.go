package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(
		&LinkCheck{},
		&LinkMonitoringJob{},
		&LinkMaintenanceReport{},
		&LinkChangeNotification{},
	)
	require.NoError(t, err)

	// Create bookmarks table for testing
	err = db.Exec(`
		CREATE TABLE bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			title TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Create bookmark_collections table for testing
	err = db.Exec(`
		CREATE TABLE bookmark_collections (
			bookmark_id INTEGER NOT NULL,
			collection_id INTEGER NOT NULL,
			PRIMARY KEY (bookmark_id, collection_id)
		)
	`).Error
	require.NoError(t, err)

	return db
}

func setupTestService(t *testing.T) (*Service, *gorm.DB) {
	db := setupTestDB(t)
	service := NewService(db)
	return service, db
}

func createTestBookmark(t *testing.T, db *gorm.DB, userID uint, url string) uint {
	result := db.Exec(`
		INSERT INTO bookmarks (user_id, url, title, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, userID, url, "Test Bookmark", time.Now(), time.Now())
	require.NoError(t, result.Error)

	var bookmarkID uint
	err := db.Raw("SELECT last_insert_rowid()").Scan(&bookmarkID).Error
	require.NoError(t, err)

	return bookmarkID
}

func TestService_CheckLink_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create test bookmark
	userID := uint(1)
	bookmarkID := createTestBookmark(t, db, userID, server.URL)

	// Test link check
	req := &CreateLinkCheckRequest{
		BookmarkID: bookmarkID,
		URL:        server.URL,
	}

	result, err := service.CheckLink(ctx, userID, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, bookmarkID, result.BookmarkID)
	assert.Equal(t, server.URL, result.URL)
	assert.Equal(t, LinkStatusActive, result.Status)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.True(t, result.ResponseTime >= 0) // Allow 0 for very fast responses
}

func TestService_CheckLink_BrokenLink(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create test bookmark
	userID := uint(1)
	bookmarkID := createTestBookmark(t, db, userID, server.URL)

	// Test link check
	req := &CreateLinkCheckRequest{
		BookmarkID: bookmarkID,
		URL:        server.URL,
	}

	result, err := service.CheckLink(ctx, userID, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, LinkStatusBroken, result.Status)
	assert.Equal(t, http.StatusNotFound, result.StatusCode)

	// Verify notification was created
	var notification LinkChangeNotification
	err = db.Where("user_id = ? AND bookmark_id = ?", userID, bookmarkID).First(&notification).Error
	require.NoError(t, err)
	assert.Equal(t, "broken", notification.ChangeType)
}

func TestService_CheckLink_Redirect(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test server that redirects
	redirectURL := "https://example.com/new-location"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", redirectURL)
			w.WriteHeader(http.StatusMovedPermanently)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create test bookmark
	userID := uint(1)
	redirectTestURL := server.URL + "/redirect"
	bookmarkID := createTestBookmark(t, db, userID, redirectTestURL)

	// Test link check
	req := &CreateLinkCheckRequest{
		BookmarkID: bookmarkID,
		URL:        redirectTestURL,
	}

	result, err := service.CheckLink(ctx, userID, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, LinkStatusRedirect, result.Status)
	assert.Equal(t, http.StatusMovedPermanently, result.StatusCode)
	assert.Equal(t, redirectURL, result.RedirectURL)
}

func TestService_CheckLink_BookmarkNotFound(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &CreateLinkCheckRequest{
		BookmarkID: 999, // Non-existent bookmark
		URL:        "https://example.com",
	}

	result, err := service.CheckLink(ctx, 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bookmark not found")
}

func TestService_CreateMonitoringJob_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &CreateMonitoringJobRequest{
		Name:        "Daily Link Check",
		Description: "Check all links daily",
		Enabled:     true,
		Frequency:   "0 0 * * *", // Daily at midnight
	}

	result, err := service.CreateMonitoringJob(ctx, 1, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, req.Enabled, result.Enabled)
	assert.Equal(t, req.Frequency, result.Frequency)
	assert.Equal(t, uint(1), result.UserID)
}

func TestService_CreateMonitoringJob_InvalidCron(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &CreateMonitoringJobRequest{
		Name:        "Invalid Job",
		Description: "Job with invalid cron",
		Enabled:     true,
		Frequency:   "invalid cron", // Invalid cron expression
	}

	result, err := service.CreateMonitoringJob(ctx, 1, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid cron expression")
}

func TestService_GetMonitoringJob_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test job
	job := &LinkMonitoringJob{
		UserID:      1,
		Name:        "Test Job",
		Description: "Test Description",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	require.NoError(t, db.Create(job).Error)

	// Get job
	result, err := service.GetMonitoringJob(ctx, 1, job.ID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, job.ID, result.ID)
	assert.Equal(t, job.Name, result.Name)
}

func TestService_GetMonitoringJob_NotFound(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	result, err := service.GetMonitoringJob(ctx, 1, 999)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "monitoring job not found")
}

func TestService_UpdateMonitoringJob_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test job
	job := &LinkMonitoringJob{
		UserID:      1,
		Name:        "Original Name",
		Description: "Original Description",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	require.NoError(t, db.Create(job).Error)

	// Update job
	enabled := false
	req := &UpdateMonitoringJobRequest{
		Name:        "Updated Name",
		Description: "Updated Description",
		Enabled:     &enabled,
		Frequency:   "0 12 * * *",
	}

	result, err := service.UpdateMonitoringJob(ctx, 1, job.ID, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, *req.Enabled, result.Enabled)
	assert.Equal(t, req.Frequency, result.Frequency)
}

func TestService_DeleteMonitoringJob_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test job
	job := &LinkMonitoringJob{
		UserID:      1,
		Name:        "Test Job",
		Description: "Test Description",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	require.NoError(t, db.Create(job).Error)

	// Delete job
	err := service.DeleteMonitoringJob(ctx, 1, job.ID)
	require.NoError(t, err)

	// Verify job is deleted
	var deletedJob LinkMonitoringJob
	err = db.Where("id = ?", job.ID).First(&deletedJob).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestService_ListMonitoringJobs_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	// Create test jobs
	for i := 0; i < 5; i++ {
		job := &LinkMonitoringJob{
			UserID:      1,
			Name:        "Test Job",
			Description: "Test Description",
			Enabled:     true,
			Frequency:   "0 0 * * *",
		}
		require.NoError(t, db.Create(job).Error)
	}

	// List jobs
	jobs, total, err := service.ListMonitoringJobs(ctx, 1, 1, 10)
	require.NoError(t, err)
	assert.Len(t, jobs, 5)
	assert.Equal(t, int64(5), total)
}

func TestService_GenerateMaintenanceReport_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	userID := uint(1)

	// Create test bookmarks
	bookmarkID1 := createTestBookmark(t, db, userID, "https://example.com/1")
	bookmarkID2 := createTestBookmark(t, db, userID, "https://example.com/2")

	// Create test link checks
	checks := []*LinkCheck{
		{
			BookmarkID: bookmarkID1,
			URL:        "https://example.com/1",
			Status:     LinkStatusActive,
			StatusCode: 200,
			CheckedAt:  time.Now(),
		},
		{
			BookmarkID: bookmarkID2,
			URL:        "https://example.com/2",
			Status:     LinkStatusBroken,
			StatusCode: 404,
			CheckedAt:  time.Now(),
		},
	}

	for _, check := range checks {
		require.NoError(t, db.Create(check).Error)
	}

	// Generate report
	report, err := service.GenerateMaintenanceReport(ctx, userID, nil)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, userID, report.UserID)
	assert.Equal(t, 2, report.TotalLinks)
	assert.Equal(t, 1, report.ActiveLinks)
	assert.Equal(t, 1, report.BrokenLinks)
	assert.Equal(t, 0, report.RedirectLinks)
	assert.NotEmpty(t, report.Suggestions)
}

func TestService_GetLinkChecks_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	userID := uint(1)
	bookmarkID := createTestBookmark(t, db, userID, "https://example.com")

	// Create test link checks
	for i := 0; i < 3; i++ {
		check := &LinkCheck{
			BookmarkID: bookmarkID,
			URL:        "https://example.com",
			Status:     LinkStatusActive,
			StatusCode: 200,
			CheckedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, db.Create(check).Error)
	}

	// Get link checks
	checks, total, err := service.GetLinkChecks(ctx, userID, bookmarkID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, checks, 3)
	assert.Equal(t, int64(3), total)
}

func TestService_GetNotifications_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	userID := uint(1)
	bookmarkID := createTestBookmark(t, db, userID, "https://example.com")

	// Create test notifications
	for i := 0; i < 3; i++ {
		notification := &LinkChangeNotification{
			UserID:     userID,
			BookmarkID: bookmarkID,
			ChangeType: "broken",
			Message:    "Link is broken",
			Read:       i == 0, // First one is read
		}
		require.NoError(t, db.Create(notification).Error)
	}

	// Get all notifications
	notifications, total, err := service.GetNotifications(ctx, userID, 1, 10, false)
	require.NoError(t, err)
	assert.Len(t, notifications, 3)
	assert.Equal(t, int64(3), total)

	// Get unread notifications only
	unreadNotifications, unreadTotal, err := service.GetNotifications(ctx, userID, 1, 10, true)
	require.NoError(t, err)
	assert.Len(t, unreadNotifications, 2)
	assert.Equal(t, int64(2), unreadTotal)
}

func TestService_MarkNotificationAsRead_Success(t *testing.T) {
	service, db := setupTestService(t)
	ctx := context.Background()

	userID := uint(1)
	bookmarkID := createTestBookmark(t, db, userID, "https://example.com")

	// Create test notification
	notification := &LinkChangeNotification{
		UserID:     userID,
		BookmarkID: bookmarkID,
		ChangeType: "broken",
		Message:    "Link is broken",
		Read:       false,
	}
	require.NoError(t, db.Create(notification).Error)

	// Mark as read
	err := service.MarkNotificationAsRead(ctx, userID, notification.ID)
	require.NoError(t, err)

	// Verify it's marked as read
	var updatedNotification LinkChangeNotification
	err = db.Where("id = ?", notification.ID).First(&updatedNotification).Error
	require.NoError(t, err)
	assert.True(t, updatedNotification.Read)
}

func TestService_IsValidCronExpression(t *testing.T) {
	service, _ := setupTestService(t)

	testCases := []struct {
		expr     string
		expected bool
	}{
		{"0 0 * * *", true},      // 5 fields - valid
		{"0 0 * * * *", true},    // 6 fields - valid
		{"0 0 * *", false},       // 4 fields - invalid
		{"0 0 * * * * *", false}, // 7 fields - invalid
		{"", false},              // empty - invalid
	}

	for _, tc := range testCases {
		result := service.isValidCronExpression(tc.expr)
		assert.Equal(t, tc.expected, result, "Expression: %s", tc.expr)
	}
}

func TestService_GenerateNotificationMessage(t *testing.T) {
	service, _ := setupTestService(t)

	testCases := []struct {
		check    *LinkCheck
		expected string
	}{
		{
			check: &LinkCheck{
				URL:        "https://example.com",
				Status:     LinkStatusBroken,
				StatusCode: 404,
			},
			expected: "Link is broken (HTTP 404): https://example.com",
		},
		{
			check: &LinkCheck{
				URL:         "https://example.com",
				Status:      LinkStatusRedirect,
				RedirectURL: "https://example.com/new",
			},
			expected: "Link redirects to: https://example.com/new",
		},
		{
			check: &LinkCheck{
				URL:    "https://example.com",
				Status: LinkStatusTimeout,
			},
			expected: "Link timed out: https://example.com",
		},
	}

	for _, tc := range testCases {
		result := service.generateNotificationMessage(tc.check)
		assert.Equal(t, tc.expected, result)
	}
}

func TestService_GenerateMaintenanceSuggestions(t *testing.T) {
	service, _ := setupTestService(t)

	testCases := []struct {
		report   *LinkMaintenanceReport
		expected []string
	}{
		{
			report: &LinkMaintenanceReport{
				TotalLinks:    10,
				ActiveLinks:   10,
				BrokenLinks:   0,
				RedirectLinks: 0,
			},
			expected: []string{"Your bookmark collection is in good health!"},
		},
		{
			report: &LinkMaintenanceReport{
				TotalLinks:    10,
				ActiveLinks:   5,
				BrokenLinks:   3,
				RedirectLinks: 2,
			},
			expected: []string{
				"Fix 3 broken links to improve collection health",
				"Update 2 redirected links to their final destinations",
				"Consider reviewing and cleaning up your bookmark collection",
			},
		},
	}

	for _, tc := range testCases {
		result := service.generateMaintenanceSuggestions(tc.report)
		assert.Equal(t, tc.expected, result)
	}
}
