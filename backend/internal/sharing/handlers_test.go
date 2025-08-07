package sharing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"bookmark-sync-service/backend/pkg/database"
)

// MockSharingService is a mock implementation of the sharing service interface
type MockSharingService struct {
	mock.Mock
}

func (m *MockSharingService) CreateShare(ctx interface{}, userID uint, request *CreateShareRequest) (*ShareResponse, error) {
	args := m.Called(ctx, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ShareResponse), args.Error(1)
}

func (m *MockSharingService) GetShareByToken(ctx interface{}, token string) (*CollectionShare, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CollectionShare), args.Error(1)
}

func (m *MockSharingService) UpdateShare(ctx interface{}, userID uint, shareID uint, request *UpdateShareRequest) (*ShareResponse, error) {
	args := m.Called(ctx, userID, shareID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ShareResponse), args.Error(1)
}

func (m *MockSharingService) DeleteShare(ctx interface{}, userID uint, shareID uint) error {
	args := m.Called(ctx, userID, shareID)
	return args.Error(0)
}

func (m *MockSharingService) GetUserShares(ctx interface{}, userID uint) ([]CollectionShare, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CollectionShare), args.Error(1)
}

func (m *MockSharingService) GetCollectionShares(ctx interface{}, userID uint, collectionID uint) ([]CollectionShare, error) {
	args := m.Called(ctx, userID, collectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CollectionShare), args.Error(1)
}

func (m *MockSharingService) ForkCollection(ctx interface{}, userID uint, originalCollectionID uint, request *ForkRequest) (*database.Collection, error) {
	args := m.Called(ctx, userID, originalCollectionID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.Collection), args.Error(1)
}

func (m *MockSharingService) AddCollaborator(ctx interface{}, userID uint, collectionID uint, request *CollaboratorRequest) (*CollectionCollaborator, error) {
	args := m.Called(ctx, userID, collectionID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CollectionCollaborator), args.Error(1)
}

func (m *MockSharingService) AcceptCollaboration(ctx interface{}, userID uint, collaboratorID uint) error {
	args := m.Called(ctx, userID, collaboratorID)
	return args.Error(0)
}

func (m *MockSharingService) RecordActivity(ctx interface{}, shareID uint, userID *uint, activityType, ipAddress, userAgent string, metadata map[string]interface{}) error {
	args := m.Called(ctx, shareID, userID, activityType, ipAddress, userAgent, metadata)
	return args.Error(0)
}

func (m *MockSharingService) GetShareActivity(ctx interface{}, userID uint, shareID uint) ([]ShareActivity, error) {
	args := m.Called(ctx, userID, shareID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ShareActivity), args.Error(1)
}

// SharingHandlerTestSuite defines the test suite for sharing handlers
type SharingHandlerTestSuite struct {
	suite.Suite
	handler     *Handler
	mockService *MockSharingService
	router      *gin.Engine
}

func (suite *SharingHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockService = new(MockSharingService)

	// Create a handler with a real service (we'll test individual functions)
	service := &Service{}
	suite.handler = NewHandler(service)

	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *SharingHandlerTestSuite) setupRoutes() {
	// Add auth middleware mock
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Set("user_email", "test@example.com")
		c.Next()
	})

	api := suite.router.Group("/api/v1")
	{
		api.POST("/shares", suite.handler.CreateShare)
		api.GET("/shares", suite.handler.GetUserShares)
		api.GET("/shared/:token", suite.handler.GetShare) // Use different path to avoid conflict
		api.PUT("/shares/:id", suite.handler.UpdateShare)
		api.DELETE("/shares/:id", suite.handler.DeleteShare)
		api.GET("/shares/:id/activity", suite.handler.GetShareActivity)
		api.GET("/collections/:id/shares", suite.handler.GetCollectionShares)
		api.POST("/collections/:id/fork", suite.handler.ForkCollection)
		api.POST("/collections/:id/collaborators", suite.handler.AddCollaborator)
		api.POST("/collaborations/:id/accept", suite.handler.AcceptCollaboration)
	}
}

func (suite *SharingHandlerTestSuite) TestCreateShareInvalidRequest() {
	request := CreateShareRequest{
		// Missing required fields
		ShareType:  ShareTypePublic,
		Permission: PermissionView,
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/shares", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *SharingHandlerTestSuite) TestGetShareMissingToken() {
	req, _ := http.NewRequest("GET", "/api/v1/shared/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusNotFound, w.Code) // Gin returns 404 for missing path params
}

func (suite *SharingHandlerTestSuite) TestUpdateShareInvalidID() {
	request := UpdateShareRequest{}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("PUT", "/api/v1/shares/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *SharingHandlerTestSuite) TestDeleteShareInvalidID() {
	req, _ := http.NewRequest("DELETE", "/api/v1/shares/invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *SharingHandlerTestSuite) TestForkCollectionInvalidID() {
	request := ForkRequest{
		Name: "Test Fork",
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/collections/invalid/fork", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *SharingHandlerTestSuite) TestAddCollaboratorInvalidID() {
	request := CollaboratorRequest{
		Email:      "test@example.com",
		Permission: PermissionEdit,
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/collections/invalid/collaborators", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *SharingHandlerTestSuite) TestAcceptCollaborationInvalidID() {
	req, _ := http.NewRequest("POST", "/api/v1/collaborations/invalid/accept", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *SharingHandlerTestSuite) TestGetShareActivityInvalidID() {
	req, _ := http.NewRequest("GET", "/api/v1/shares/invalid/activity", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

// Run the test suite
func TestSharingHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SharingHandlerTestSuite))
}

// Test individual handler functions
func TestNewHandler(t *testing.T) {
	service := &Service{}
	handler := NewHandler(service)

	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.service)
}

// Test handler with no authentication
func TestCreateShareNoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	service := &Service{}
	handler := NewHandler(service)

	router.POST("/shares", handler.CreateShare)

	request := CreateShareRequest{
		CollectionID: 1,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/shares", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test handler with invalid user ID
func TestCreateShareInvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock middleware that sets invalid user ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "invalid")
		c.Next()
	})

	service := &Service{}
	handler := NewHandler(service)

	router.POST("/shares", handler.CreateShare)

	request := CreateShareRequest{
		CollectionID: 1,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/shares", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
