package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/internal/user"
	"bookmark-sync-service/backend/pkg/validation"
	"bookmark-sync-service/backend/pkg/worker"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestJob implements the Job interface for testing
type TestJob struct {
	worker.BaseJob
	ExecuteFunc func(ctx context.Context) error
	executed    bool
	mu          sync.Mutex
}

func NewTestJob(id, jobType string, maxRetries int, executeFunc func(ctx context.Context) error) *TestJob {
	return &TestJob{
		BaseJob: worker.BaseJob{
			ID:         id,
			Type:       jobType,
			MaxRetries: maxRetries,
			CreatedAt:  time.Now(),
		},
		ExecuteFunc: executeFunc,
	}
}

func (j *TestJob) Execute(ctx context.Context) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.executed = true
	if j.ExecuteFunc != nil {
		return j.ExecuteFunc(ctx)
	}
	return nil
}

func (j *TestJob) IsExecuted() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.executed
}

// MockUserService implements user.ServiceInterface for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uint) (*user.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserProfile), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint, req *user.UpdateProfileRequest) (*user.UserProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserProfile), args.Error(1)
}

func (m *MockUserService) UpdatePreferences(ctx context.Context, userID uint, req *user.UpdatePreferencesRequest) (*user.UserProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserProfile), args.Error(1)
}

func (m *MockUserService) UploadAvatar(ctx context.Context, userID uint, imageData []byte, contentType string) (*user.UserProfile, error) {
	args := m.Called(ctx, userID, imageData, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserProfile), args.Error(1)
}

func (m *MockUserService) ExportUserData(ctx context.Context, userID uint) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type IntegrationTestSuite struct {
	suite.Suite
	router      *gin.Engine
	userService *MockUserService
	workerPool  *worker.WorkerPool
	logger      *zap.Logger
}

func (suite *IntegrationTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.logger = zaptest.NewLogger(suite.T())
	suite.userService = new(MockUserService)
	suite.workerPool = worker.NewWorkerPool(2, 10, suite.logger)
	suite.workerPool.Start()

	suite.router = gin.New()

	// Add middleware to simulate authentication
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", "123")
		c.Next()
	})

	// Setup user routes
	userHandler := user.NewHandler(suite.userService, suite.logger)
	userGroup := suite.router.Group("/api/v1/users")
	{
		userGroup.GET("/profile", userHandler.GetProfile)
		userGroup.PUT("/profile", userHandler.UpdateProfile)
	}
}

func (suite *IntegrationTestSuite) TearDownTest() {
	suite.workerPool.Stop()
}

func (suite *IntegrationTestSuite) TestConstants_AreUsedCorrectly() {
	// Test that constants are properly defined and accessible
	assert.Equal(suite.T(), 24*time.Hour, config.DefaultCacheTTL)
	assert.Equal(suite.T(), 5*time.Second, config.DefaultConnectionTimeout)
	assert.Equal(suite.T(), 200, config.DefaultThumbnailSize)
	assert.Equal(suite.T(), 20, config.DefaultPageSize)
	assert.Equal(suite.T(), 100, config.MaxPageSize)
}

func (suite *IntegrationTestSuite) TestValidation_UserIDFromContext() {
	validator := validation.NewRequestValidator()

	// Create a test context with user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "123")

	userID, err := validator.UserIDFromContext(c)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint(123), userID)
}

func (suite *IntegrationTestSuite) TestValidation_PaginationParams() {
	validator := validation.NewRequestValidator()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?page=2&page_size=10", nil)

	params, err := validator.ValidatePagination(c)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, params.Page)
	assert.Equal(suite.T(), 10, params.PageSize)
	assert.Equal(suite.T(), 10, params.Offset) // (2-1) * 10
}

func (suite *IntegrationTestSuite) TestWorkerPool_JobExecution() {
	executed := false
	job := NewTestJob("test-1", "test", 3, func(ctx context.Context) error {
		executed = true
		return nil
	})

	err := suite.workerPool.Submit(job)
	assert.NoError(suite.T(), err)

	// Wait for job to be processed
	time.Sleep(100 * time.Millisecond)
	assert.True(suite.T(), executed)
}

func (suite *IntegrationTestSuite) TestUserHandler_GetProfile_Success() {
	// Mock service response
	expectedProfile := &user.UserProfile{
		ID:       123,
		Username: "testuser",
		Email:    "test@example.com",
	}
	suite.userService.On("GetProfile", mock.Anything, uint(123)).Return(expectedProfile, nil)

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Check if response has the expected structure
	if response["status"] != nil {
		assert.Equal(suite.T(), "success", response["status"])
	} else {
		// If no status field, check if we have the profile data directly
		assert.NotNil(suite.T(), response["data"])
	}

	suite.userService.AssertExpectations(suite.T())
}

func (suite *IntegrationTestSuite) TestUserHandler_UpdateProfile_ValidationError() {
	// Invalid request data (username too short - min=3 required)
	invalidData := map[string]interface{}{
		"username": "ab", // Username too short, should fail validation
	}
	jsonData, _ := json.Marshal(invalidData)

	req := httptest.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Should return validation error (400 for validation errors)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), false, response["success"])

	// Check the nested error structure
	errorObj, exists := response["error"].(map[string]interface{})
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), "VALIDATION_ERROR", errorObj["code"])

	// Check for validation_errors in details
	details, exists := errorObj["details"].(map[string]interface{})
	assert.True(suite.T(), exists)
	assert.Contains(suite.T(), details, "validation_errors")
}

func (suite *IntegrationTestSuite) TestErrorHandling_ConsistentResponses() {
	validator := validation.NewRequestValidator()

	// Test unauthorized error
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	validator.HandleUnauthorizedError(c, config.ErrUserNotAuthenticated)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test not found error
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	validator.HandleNotFoundError(c, "User")

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	// Test internal error
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	validator.HandleInternalError(c, config.ErrInternalError)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

func (suite *IntegrationTestSuite) TestWorkerPool_GracefulShutdown() {
	// Create a new worker pool for this test
	pool := worker.NewWorkerPool(1, 5, suite.logger)
	pool.Start()

	// Submit a job
	job := NewTestJob("shutdown-test", "test", 3, func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond) // Simulate work
		return nil
	})

	err := pool.Submit(job)
	assert.NoError(suite.T(), err)

	// Give the job time to start processing
	time.Sleep(10 * time.Millisecond)

	// Stop the pool (should wait for job to complete)
	pool.Stop()

	// Job should have completed
	assert.True(suite.T(), job.IsExecuted())
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
