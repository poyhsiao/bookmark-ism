package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *Handler, *gorm.DB) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
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

	// Create service and handler
	service := NewService(db)
	handler := NewHandler(service)

	// Setup router
	router := gin.New()

	// Add middleware to set user ID in context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1") // Set as string, as middleware does
		c.Next()
	})

	// Register routes
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	return router, handler, db
}

func createTestBookmarkForHandler(t *testing.T, db *gorm.DB, userID uint, url string) uint {
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

func TestHandler_CheckLink_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test bookmark
	bookmarkID := createTestBookmarkForHandler(t, db, 1, "https://example.com")

	// Create request
	reqBody := CreateLinkCheckRequest{
		BookmarkID: bookmarkID,
		URL:        "https://example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/monitoring/check-link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response LinkCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.LinkCheck)
	assert.Equal(t, bookmarkID, response.LinkCheck.BookmarkID)
	assert.Equal(t, "https://example.com", response.LinkCheck.URL)
}

func TestHandler_CheckLink_InvalidRequest(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create invalid request (missing required fields)
	reqBody := map[string]interface{}{
		"url": "invalid-url",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/monitoring/check-link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CheckLink_BookmarkNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create request with non-existent bookmark
	reqBody := CreateLinkCheckRequest{
		BookmarkID: 999,
		URL:        "https://example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/monitoring/check-link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_GetLinkChecks_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test bookmark
	bookmarkID := createTestBookmarkForHandler(t, db, 1, "https://example.com")

	// Create test link checks
	for i := 0; i < 3; i++ {
		check := &LinkCheck{
			BookmarkID: bookmarkID,
			URL:        "https://example.com",
			Status:     LinkStatusActive,
			StatusCode: 200,
			CheckedAt:  time.Now(),
		}
		require.NoError(t, db.Create(check).Error)
	}

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/monitoring/bookmarks/%d/checks", bookmarkID), nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response ListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(3), response.Total)
}

func TestHandler_GetLinkChecks_InvalidBookmarkID(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Make request with invalid bookmark ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/monitoring/bookmarks/invalid/checks", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateMonitoringJob_Success(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create request
	reqBody := CreateMonitoringJobRequest{
		Name:        "Daily Check",
		Description: "Check all links daily",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/monitoring/jobs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response MonitoringJobResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Job)
	assert.Equal(t, reqBody.Name, response.Job.Name)
}

func TestHandler_CreateMonitoringJob_InvalidCron(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create request with invalid cron
	reqBody := CreateMonitoringJobRequest{
		Name:        "Invalid Job",
		Description: "Job with invalid cron",
		Enabled:     true,
		Frequency:   "invalid",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/monitoring/jobs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_ListMonitoringJobs_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test jobs
	for i := 0; i < 3; i++ {
		job := &LinkMonitoringJob{
			UserID:      1,
			Name:        fmt.Sprintf("Job %d", i),
			Description: "Test job",
			Enabled:     true,
			Frequency:   "0 0 * * *",
		}
		require.NoError(t, db.Create(job).Error)
	}

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/monitoring/jobs", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response ListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(3), response.Total)
}

func TestHandler_GetMonitoringJob_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test job
	job := &LinkMonitoringJob{
		UserID:      1,
		Name:        "Test Job",
		Description: "Test Description",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	require.NoError(t, db.Create(job).Error)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/monitoring/jobs/%d", job.ID), nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response MonitoringJobResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Job)
	assert.Equal(t, job.ID, response.Job.ID)
}

func TestHandler_GetMonitoringJob_NotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Make request with non-existent job ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/monitoring/jobs/999", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdateMonitoringJob_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test job
	job := &LinkMonitoringJob{
		UserID:      1,
		Name:        "Original Name",
		Description: "Original Description",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	require.NoError(t, db.Create(job).Error)

	// Create update request
	enabled := false
	reqBody := UpdateMonitoringJobRequest{
		Name:        "Updated Name",
		Description: "Updated Description",
		Enabled:     &enabled,
		Frequency:   "0 12 * * *",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/monitoring/jobs/%d", job.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response MonitoringJobResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Job)
	assert.Equal(t, reqBody.Name, response.Job.Name)
	assert.Equal(t, *reqBody.Enabled, response.Job.Enabled)
}

func TestHandler_DeleteMonitoringJob_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test job
	job := &LinkMonitoringJob{
		UserID:      1,
		Name:        "Test Job",
		Description: "Test Description",
		Enabled:     true,
		Frequency:   "0 0 * * *",
	}
	require.NoError(t, db.Create(job).Error)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/monitoring/jobs/%d", job.ID), nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify job is deleted
	var deletedJob LinkMonitoringJob
	err := db.Where("id = ?", job.ID).First(&deletedJob).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestHandler_GenerateMaintenanceReport_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test bookmarks
	bookmarkID1 := createTestBookmarkForHandler(t, db, 1, "https://example.com/1")
	bookmarkID2 := createTestBookmarkForHandler(t, db, 1, "https://example.com/2")

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

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/monitoring/reports", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response MaintenanceReportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotNil(t, response.Report)
	assert.Equal(t, 2, response.Report.TotalLinks)
}

func TestHandler_GetNotifications_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test bookmark
	bookmarkID := createTestBookmarkForHandler(t, db, 1, "https://example.com")

	// Create test notifications
	for i := 0; i < 3; i++ {
		notification := &LinkChangeNotification{
			UserID:     1,
			BookmarkID: bookmarkID,
			ChangeType: "broken",
			Message:    "Link is broken",
			Read:       i == 0, // First one is read
		}
		require.NoError(t, db.Create(notification).Error)
	}

	// Make request for all notifications
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/monitoring/notifications", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response ListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(3), response.Total)

	// Make request for unread notifications only
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/monitoring/notifications?unread_only=true", nil)
	router.ServeHTTP(w2, req2)

	// Assert response
	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 ListResponse
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), response2.Total)
}

func TestHandler_MarkNotificationAsRead_Success(t *testing.T) {
	router, _, db := setupTestRouter(t)

	// Create test bookmark
	bookmarkID := createTestBookmarkForHandler(t, db, 1, "https://example.com")

	// Create test notification
	notification := &LinkChangeNotification{
		UserID:     1,
		BookmarkID: bookmarkID,
		ChangeType: "broken",
		Message:    "Link is broken",
		Read:       false,
	}
	require.NoError(t, db.Create(notification).Error)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/monitoring/notifications/%d/read", notification.ID), nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify notification is marked as read
	var updatedNotification LinkChangeNotification
	err := db.Where("id = ?", notification.ID).First(&updatedNotification).Error
	require.NoError(t, err)
	assert.True(t, updatedNotification.Read)
}

func TestHandler_MarkNotificationAsRead_NotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Make request with non-existent notification ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/monitoring/notifications/999/read", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Unauthorized(t *testing.T) {
	// Setup router without user ID in context
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	service := NewService(db)
	handler := NewHandler(service)

	router := gin.New()
	// No middleware to set user ID - should result in unauthorized

	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	// Test endpoints that require authentication
	endpoints := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"POST", "/api/v1/monitoring/check-link", CreateLinkCheckRequest{BookmarkID: 1, URL: "https://example.com"}},
		{"GET", "/api/v1/monitoring/bookmarks/1/checks", nil},
		{"POST", "/api/v1/monitoring/jobs", CreateMonitoringJobRequest{Name: "Test", Frequency: "0 0 * * *"}},
		{"GET", "/api/v1/monitoring/jobs", nil},
		{"GET", "/api/v1/monitoring/jobs/1", nil},
		{"PUT", "/api/v1/monitoring/jobs/1", UpdateMonitoringJobRequest{Name: "Updated"}},
		{"DELETE", "/api/v1/monitoring/jobs/1", nil},
		{"POST", "/api/v1/monitoring/reports", nil},
		{"GET", "/api/v1/monitoring/notifications", nil},
		{"PUT", "/api/v1/monitoring/notifications/1/read", nil},
	}

	for _, endpoint := range endpoints {
		var req *http.Request
		if endpoint.body != nil {
			jsonBody, _ := json.Marshal(endpoint.body)
			req, _ = http.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, _ = http.NewRequest(endpoint.method, endpoint.path, nil)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Endpoint: %s %s", endpoint.method, endpoint.path)
	}
}
