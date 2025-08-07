package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockUserService is a mock implementation of the user service
// MockUserService 是用戶服務的模擬實現
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uint) (*UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserProfile), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*UserProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserProfile), args.Error(1)
}

func (m *MockUserService) UpdatePreferences(ctx context.Context, userID uint, req *UpdatePreferencesRequest) (*UserProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserProfile), args.Error(1)
}

func (m *MockUserService) UploadAvatar(ctx context.Context, userID uint, imageData []byte, contentType string) (*UserProfile, error) {
	args := m.Called(ctx, userID, imageData, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserProfile), args.Error(1)
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

// setupTestHandler creates a test handler with mock service
// setupTestHandler 創建帶有模擬服務的測試處理器
func setupTestHandler() (*Handler, *MockUserService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockUserService)
	logger := zap.NewNop()
	handler := NewHandler(mockService, logger)
	return handler, mockService
}

// setupTestRouter creates a test router with the handler
// setupTestRouter 創建帶有處理器的測試路由器
func setupTestRouter(handler *Handler) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("trace", &utils.TraceContext{
			Logger: zap.NewNop(),
		})
		c.Next()
	})
	return router
}

// TestGetProfile tests the GetProfile handler
// TestGetProfile 測試 GetProfile 處理器
func TestGetProfile(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.GET("/profile", handler.GetProfile)

	t.Run("Get Profile Successfully", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockProfile := &UserProfile{
			ID:          1,
			Email:       "test@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
			Preferences: UserPreferences{
				Theme:       "light",
				GridSize:    "medium",
				DefaultView: "grid",
			},
		}
		mockService.On("GetProfile", mock.Anything, uint(1)).Return(mockProfile, nil).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusOK, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Profile retrieved successfully")

		mockService.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Mock service response for non-existent user
		// 模擬不存在用戶的服務響應
		mockService.On("GetProfile", mock.Anything, uint(1)).Return(nil, errors.New("user not found")).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "NOT_FOUND", response.Error.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		// Mock service response for internal error
		// 模擬內部錯誤的服務響應
		mockService.On("GetProfile", mock.Anything, uint(1)).Return(nil, errors.New("database error")).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)

		mockService.AssertExpectations(t)
	})
}

// TestUpdateProfile tests the UpdateProfile handler
// TestUpdateProfile 測試 UpdateProfile 處理器
func TestUpdateProfile(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.PUT("/profile", handler.UpdateProfile)

	t.Run("Update Profile Successfully", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockProfile := &UserProfile{
			ID:          1,
			Email:       "test@example.com",
			Username:    "updateduser",
			DisplayName: "Updated User",
		}
		mockService.On("UpdateProfile", mock.Anything, uint(1), mock.AnythingOfType("*user.UpdateProfileRequest")).
			Return(mockProfile, nil).Once()

		// Create request
		// 創建請求
		reqBody := UpdateProfileRequest{
			DisplayName: "Updated User",
			Username:    "updateduser",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusOK, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Profile updated successfully")

		mockService.AssertExpectations(t)
	})

	t.Run("Username Already Taken", func(t *testing.T) {
		// Mock service response for username conflict
		// 模擬用戶名衝突的服務響應
		mockService.On("UpdateProfile", mock.Anything, uint(1), mock.AnythingOfType("*user.UpdateProfileRequest")).
			Return(nil, errors.New("username already taken")).Once()

		// Create request
		// 創建請求
		reqBody := UpdateProfileRequest{
			Username: "takenusername",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusConflict, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "USERNAME_TAKEN", response.Error.Code)

		mockService.AssertExpectations(t)
	})
}

// TestUpdatePreferences tests the UpdatePreferences handler
// TestUpdatePreferences 測試 UpdatePreferences 處理器
func TestUpdatePreferences(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.PUT("/preferences", handler.UpdatePreferences)

	t.Run("Update Preferences Successfully", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockProfile := &UserProfile{
			ID:       1,
			Email:    "test@example.com",
			Username: "testuser",
			Preferences: UserPreferences{
				Theme:       "dark",
				GridSize:    "large",
				DefaultView: "list",
				Language:    "zh-CN",
			},
		}
		mockService.On("UpdatePreferences", mock.Anything, uint(1), mock.AnythingOfType("*user.UpdatePreferencesRequest")).
			Return(mockProfile, nil).Once()

		// Create request
		// 創建請求
		reqBody := UpdatePreferencesRequest{
			Theme:       "dark",
			GridSize:    "large",
			DefaultView: "list",
			Language:    "zh-CN",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/preferences", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusOK, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Preferences updated successfully")

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Preferences", func(t *testing.T) {
		// Create request with invalid preferences
		// 創建帶有無效偏好設置的請求
		reqBody := `{"theme": "invalid-theme"}`
		req, _ := http.NewRequest("PUT", "/preferences", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})
}

// TestGetStats tests the GetStats handler
// TestGetStats 測試 GetStats 處理器
func TestGetStats(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.GET("/stats", handler.GetStats)

	t.Run("Get Stats Successfully", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockProfile := &UserProfile{
			ID:       1,
			Email:    "test@example.com",
			Username: "testuser",
			Stats: UserStats{
				BookmarkCount:   10,
				CollectionCount: 3,
				StorageUsed:     1024,
			},
		}
		mockService.On("GetProfile", mock.Anything, uint(1)).Return(mockProfile, nil).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("GET", "/stats", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusOK, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Stats retrieved successfully")

		// Check stats data
		// 檢查統計數據
		data := response.Data.(map[string]interface{})
		assert.Equal(t, float64(10), data["bookmark_count"])
		assert.Equal(t, float64(3), data["collection_count"])
		assert.Equal(t, float64(1024), data["storage_used"])

		mockService.AssertExpectations(t)
	})
}

// TestExportData tests the ExportData handler
// TestExportData 測試 ExportData 處理器
func TestExportData(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.POST("/export", handler.ExportData)

	t.Run("Export Data Successfully", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockExportData := map[string]interface{}{
			"user_profile": map[string]interface{}{
				"id":       1,
				"email":    "test@example.com",
				"username": "testuser",
			},
			"bookmarks":   []interface{}{},
			"collections": []interface{}{},
			"statistics": map[string]interface{}{
				"bookmark_count":   10,
				"collection_count": 3,
			},
		}
		mockService.On("ExportUserData", mock.Anything, uint(1)).Return(mockExportData, nil).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("POST", "/export", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusOK, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Data exported successfully")

		mockService.AssertExpectations(t)
	})
}

// TestDeleteAccount tests the DeleteAccount handler
// TestDeleteAccount 測試 DeleteAccount 處理器
func TestDeleteAccount(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.DELETE("/account", handler.DeleteAccount)

	t.Run("Delete Account Successfully", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockService.On("DeleteUser", mock.Anything, uint(1)).Return(nil).Once()

		// Create request with confirmation
		// 創建帶有確認的請求
		req, _ := http.NewRequest("DELETE", "/account?confirm=DELETE_MY_ACCOUNT", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusOK, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Account deleted successfully")

		mockService.AssertExpectations(t)
	})

	t.Run("Delete Account Without Confirmation", func(t *testing.T) {
		// Create request without confirmation
		// 創建沒有確認的請求
		req, _ := http.NewRequest("DELETE", "/account", nil)
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "CONFIRMATION_REQUIRED", response.Error.Code)
	})
}
