package offline

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockOfflineService for testing handlers
type MockOfflineService struct {
	mock.Mock
}

func (m *MockOfflineService) CacheBookmark(ctx context.Context, bookmark *database.Bookmark) error {
	args := m.Called(ctx, bookmark)
	return args.Error(0)
}

func (m *MockOfflineService) GetCachedBookmark(ctx context.Context, userID, bookmarkID uint) (*database.Bookmark, error) {
	args := m.Called(ctx, userID, bookmarkID)
	return args.Get(0).(*database.Bookmark), args.Error(1)
}

func (m *MockOfflineService) CacheBookmarks(ctx context.Context, bookmarks []database.Bookmark) error {
	args := m.Called(ctx, bookmarks)
	return args.Error(0)
}

func (m *MockOfflineService) GetCachedBookmarksForUser(ctx context.Context, userID uint) ([]database.Bookmark, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]database.Bookmark), args.Error(1)
}

func (m *MockOfflineService) QueueOfflineChange(ctx context.Context, change *OfflineChange) error {
	args := m.Called(ctx, change)
	return args.Error(0)
}

func (m *MockOfflineService) GetOfflineQueue(ctx context.Context, userID uint) ([]*OfflineChange, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*OfflineChange), args.Error(1)
}

func (m *MockOfflineService) ProcessOfflineQueue(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockOfflineService) CheckConnectivity(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockOfflineService) GetOfflineStatus(ctx context.Context, userID uint) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockOfflineService) SetOfflineStatus(ctx context.Context, userID uint, status string) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockOfflineService) CleanupExpiredCache(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockOfflineService) GetCacheStats(ctx context.Context, userID uint) (*CacheStats, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*CacheStats), args.Error(1)
}

func (m *MockOfflineService) UpdateCacheStats(ctx context.Context, userID uint, stats *CacheStats) error {
	args := m.Called(ctx, userID, stats)
	return args.Error(0)
}

func (m *MockOfflineService) ResolveConflict(localChange, serverChange *OfflineChange) *OfflineChange {
	args := m.Called(localChange, serverChange)
	return args.Get(0).(*OfflineChange)
}

func (m *MockOfflineService) SyncWhenOnline(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockOfflineService) CreateOfflineChange(userID uint, deviceID, changeType, resourceID, data string) *OfflineChange {
	args := m.Called(userID, deviceID, changeType, resourceID, data)
	return args.Get(0).(*OfflineChange)
}

func (m *MockOfflineService) ValidateChangeType(changeType string) bool {
	args := m.Called(changeType)
	return args.Bool(0)
}

func (m *MockOfflineService) GetOfflineIndicator(ctx context.Context, userID uint) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// OfflineHandlersTestSuite defines the test suite for offline handlers
type OfflineHandlersTestSuite struct {
	suite.Suite
	router  *gin.Engine
	service *MockOfflineService
	handler *Handler
}

func (suite *OfflineHandlersTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.service = &MockOfflineService{}
	suite.handler = NewHandler(suite.service)

	// Setup routes
	api := suite.router.Group("/api/v1")
	{
		offline := api.Group("/offline")
		{
			offline.POST("/cache/bookmark", suite.handler.CacheBookmark)
			offline.GET("/cache/bookmark/:id", suite.handler.GetCachedBookmark)
			offline.GET("/cache/bookmarks", suite.handler.GetCachedBookmarksForUser)
			offline.POST("/queue/change", suite.handler.QueueOfflineChange)
			offline.GET("/queue", suite.handler.GetOfflineQueue)
			offline.POST("/sync", suite.handler.ProcessOfflineQueue)
			offline.GET("/status", suite.handler.GetOfflineStatus)
			offline.PUT("/status", suite.handler.SetOfflineStatus)
			offline.GET("/stats", suite.handler.GetCacheStats)
			offline.DELETE("/cache/cleanup", suite.handler.CleanupExpiredCache)
			offline.GET("/indicator", suite.handler.GetOfflineIndicator)
			offline.GET("/connectivity", suite.handler.CheckConnectivity)
		}
	}
}

func (suite *OfflineHandlersTestSuite) TestCacheBookmark() {
	bookmark := &database.Bookmark{
		BaseModel:   database.BaseModel{ID: 1},
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Example",
		Description: "Test bookmark",
		Tags:        `["test"]`,
	}

	// Mock service call
	suite.service.On("CacheBookmark", mock.Anything, mock.AnythingOfType("*database.Bookmark")).Return(nil)

	bookmarkJSON, _ := json.Marshal(bookmark)
	req, _ := http.NewRequest("POST", "/api/v1/offline/cache/bookmark", bytes.NewBuffer(bookmarkJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestGetCachedBookmark() {
	bookmark := &database.Bookmark{
		BaseModel:   database.BaseModel{ID: 1},
		UserID:      1,
		URL:         "https://example.com",
		Title:       "Example",
		Description: "Test bookmark",
		Tags:        `["test"]`,
	}

	// Mock service call
	suite.service.On("GetCachedBookmark", mock.Anything, uint(1), uint(1)).Return(bookmark, nil)

	req, _ := http.NewRequest("GET", "/api/v1/offline/cache/bookmark/1", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestGetCachedBookmarkNotFound() {
	// Mock service call - bookmark not found
	suite.service.On("GetCachedBookmark", mock.Anything, uint(1), uint(999)).Return((*database.Bookmark)(nil), ErrBookmarkNotCached)

	req, _ := http.NewRequest("GET", "/api/v1/offline/cache/bookmark/999", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestQueueOfflineChange() {
	changeData := map[string]interface{}{
		"device_id":   "device-123",
		"type":        "bookmark_create",
		"resource_id": "bookmark-1",
		"data":        `{"url":"https://example.com","title":"Example"}`,
	}

	// Mock service call
	suite.service.On("ValidateChangeType", "bookmark_create").Return(true)
	suite.service.On("QueueOfflineChange", mock.Anything, mock.AnythingOfType("*offline.OfflineChange")).Return(nil)

	changeJSON, _ := json.Marshal(changeData)
	req, _ := http.NewRequest("POST", "/api/v1/offline/queue/change", bytes.NewBuffer(changeJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestQueueOfflineChangeInvalidType() {
	changeData := map[string]interface{}{
		"device_id":   "device-123",
		"type":        "invalid_type",
		"resource_id": "bookmark-1",
		"data":        `{"url":"https://example.com","title":"Example"}`,
	}

	// Mock service call
	suite.service.On("ValidateChangeType", "invalid_type").Return(false)

	changeJSON, _ := json.Marshal(changeData)
	req, _ := http.NewRequest("POST", "/api/v1/offline/queue/change", bytes.NewBuffer(changeJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestGetOfflineQueue() {
	changes := []*OfflineChange{
		{
			ID:         "change-1",
			UserID:     1,
			DeviceID:   "device-123",
			Type:       "bookmark_create",
			ResourceID: "bookmark-1",
			Data:       `{"url":"https://example.com","title":"Example"}`,
			Timestamp:  time.Now(),
		},
	}

	// Mock service call
	suite.service.On("GetOfflineQueue", mock.Anything, uint(1)).Return(changes, nil)

	req, _ := http.NewRequest("GET", "/api/v1/offline/queue", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestProcessOfflineQueue() {
	// Mock service call
	suite.service.On("ProcessOfflineQueue", mock.Anything, uint(1)).Return(nil)

	req, _ := http.NewRequest("POST", "/api/v1/offline/sync", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestGetOfflineStatus() {
	// Mock service call
	suite.service.On("GetOfflineStatus", mock.Anything, uint(1)).Return("offline", nil)

	req, _ := http.NewRequest("GET", "/api/v1/offline/status", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestSetOfflineStatus() {
	statusData := map[string]string{
		"status": "offline",
	}

	// Mock service call
	suite.service.On("SetOfflineStatus", mock.Anything, uint(1), "offline").Return(nil)

	statusJSON, _ := json.Marshal(statusData)
	req, _ := http.NewRequest("PUT", "/api/v1/offline/status", bytes.NewBuffer(statusJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestGetCacheStats() {
	stats := &CacheStats{
		CachedBookmarksCount: 10,
		QueuedChangesCount:   5,
		LastSync:             time.Now(),
		CacheSize:            1024,
	}

	// Mock service call
	suite.service.On("GetCacheStats", mock.Anything, uint(1)).Return(stats, nil)

	req, _ := http.NewRequest("GET", "/api/v1/offline/stats", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestCleanupExpiredCache() {
	// Mock service call
	suite.service.On("CleanupExpiredCache", mock.Anything, uint(1)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/offline/cache/cleanup", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestGetOfflineIndicator() {
	indicator := map[string]interface{}{
		"status":           "offline",
		"is_online":        false,
		"cached_bookmarks": 10,
		"queued_changes":   5,
		"last_sync":        time.Now(),
	}

	// Mock service call
	suite.service.On("GetOfflineIndicator", mock.Anything, uint(1)).Return(indicator, nil)

	req, _ := http.NewRequest("GET", "/api/v1/offline/indicator", nil)
	req.Header.Set("X-User-ID", "1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.service.AssertExpectations(suite.T())
}

func (suite *OfflineHandlersTestSuite) TestCheckConnectivity() {
	// Mock service call
	suite.service.On("CheckConnectivity", mock.Anything).Return(true)

	req, _ := http.NewRequest("GET", "/api/v1/offline/connectivity", nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.service.AssertExpectations(suite.T())
}

// Run the test suite
func TestOfflineHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(OfflineHandlersTestSuite))
}

// Additional unit tests
func TestNewHandler(t *testing.T) {
	service := &MockOfflineService{}
	handler := NewHandler(service)

	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.service)
}
