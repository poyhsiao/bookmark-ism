package auth

import (
	"bytes"
	"encoding/json"
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

// MockAuthService is a mock implementation of the auth service
// MockAuthService 是認證服務的模擬實現
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx interface{}, req *RegisterRequest) (*AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx interface{}, req *LoginRequest) (*AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx interface{}, req *RefreshRequest) (*AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func (m *MockAuthService) Logout(ctx interface{}, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) ResetPassword(ctx interface{}, req *ResetPasswordRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*UserInfo, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserInfo), args.Error(1)
}

// setupTestHandler creates a test handler with mock service
// setupTestHandler 創建帶有模擬服務的測試處理器
func setupTestHandler() (*Handler, *MockAuthService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
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

// TestRegister tests the Register handler
// TestRegister 測試 Register 處理器
func TestRegister(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)
	router.POST("/register", handler.Register)

	t.Run("Successful Registration", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockResponse := &AuthResponse{
			User: &UserInfo{
				ID:          1,
				Email:       "test@example.com",
				Username:    "testuser",
				DisplayName: "Test User",
			},
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
		}
		mockService.On("Register", mock.Anything, mock.AnythingOfType("*auth.RegisterRequest")).Return(mockResponse, nil).Once()

		// Create request
		// 創建請求
		reqBody := RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			Username:    "testuser",
			DisplayName: "Test User",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
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
		assert.Contains(t, response.Message, "registered")

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		// Create invalid request (missing required fields)
		// 創建無效請求（缺少必填字段）
		reqBody := `{"email": "test@example.com"}`
		req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
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

	t.Run("User Already Exists", func(t *testing.T) {
		// Mock service response for existing user
		// 模擬現有用戶的服務響應
		mockService.On("Register", mock.Anything, mock.AnythingOfType("*auth.RegisterRequest")).
			Return(nil, assert.AnError).Once()

		// Create request
		// 創建請求
		reqBody := RegisterRequest{
			Email:       "existing@example.com",
			Password:    "password123",
			Username:    "existinguser",
			DisplayName: "Existing User",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
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

		mockService.AssertExpectations(t)
	})
}

// TestLogin tests the Login handler
// TestLogin 測試 Login 處理器
func TestLogin(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)
	router.POST("/login", handler.Login)

	t.Run("Successful Login", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockResponse := &AuthResponse{
			User: &UserInfo{
				ID:          1,
				Email:       "test@example.com",
				Username:    "testuser",
				DisplayName: "Test User",
			},
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresIn:    3600,
		}
		mockService.On("Login", mock.Anything, mock.AnythingOfType("*auth.LoginRequest")).Return(mockResponse, nil).Once()

		// Create request
		// 創建請求
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
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
		assert.Contains(t, response.Message, "Login successful")

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		// Mock service response for invalid credentials
		// 模擬無效憑據的服務響應
		mockService.On("Login", mock.Anything, mock.AnythingOfType("*auth.LoginRequest")).
			Return(nil, assert.AnError).Once()

		// Create request
		// 創建請求
		reqBody := LoginRequest{
			Email:    "wrong@example.com",
			Password: "wrongpassword",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
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

		mockService.AssertExpectations(t)
	})
}

// TestRefreshToken tests the RefreshToken handler
// TestRefreshToken 測試 RefreshToken 處理器
func TestRefreshToken(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)
	router.POST("/refresh", handler.RefreshToken)

	t.Run("Successful Token Refresh", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockResponse := &AuthResponse{
			User: &UserInfo{
				ID:          1,
				Email:       "test@example.com",
				Username:    "testuser",
				DisplayName: "Test User",
			},
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
		}
		mockService.On("RefreshToken", mock.Anything, mock.AnythingOfType("*auth.RefreshRequest")).Return(mockResponse, nil).Once()

		// Create request
		// 創建請求
		reqBody := RefreshRequest{
			RefreshToken: "valid-refresh-token",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
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
		assert.Contains(t, response.Message, "Token refreshed")

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		// Mock service response for invalid refresh token
		// 模擬無效刷新令牌的服務響應
		mockService.On("RefreshToken", mock.Anything, mock.AnythingOfType("*auth.RefreshRequest")).
			Return(nil, assert.AnError).Once()

		// Create request
		// 創建請求
		reqBody := RefreshRequest{
			RefreshToken: "invalid-refresh-token",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "INVALID_REFRESH_TOKEN", response.Error.Code)

		mockService.AssertExpectations(t)
	})
}

// TestLogout tests the Logout handler
// TestLogout 測試 Logout 處理器
func TestLogout(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)

	// Add middleware to set user ID
	// 添加中間件來設置用戶 ID
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "1")
		c.Next()
	})

	router.POST("/logout", handler.Logout)

	t.Run("Successful Logout", func(t *testing.T) {
		// Mock service response
		// 模擬服務���應
		mockService.On("Logout", mock.Anything, uint(1)).Return(nil).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("POST", "/logout", nil)
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
		assert.Contains(t, response.Message, "Logout successful")

		mockService.AssertExpectations(t)
	})
}

// TestResetPassword tests the ResetPassword handler
// TestResetPassword 測試 ResetPassword 處理器
func TestResetPassword(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)
	router.POST("/reset", handler.ResetPassword)

	t.Run("Successful Password Reset Request", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockService.On("ResetPassword", mock.Anything, mock.AnythingOfType("*auth.ResetPasswordRequest")).Return(nil).Once()

		// Create request
		// 創建請求
		reqBody := ResetPasswordRequest{
			Email: "test@example.com",
		}
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/reset", bytes.NewBuffer(jsonData))
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
		assert.Contains(t, response.Message, "Password reset email sent")

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		// Create request with invalid email
		// 創建帶有無效電子郵件的請求
		reqBody := `{"email": "not-an-email"}`
		req, _ := http.NewRequest("POST", "/reset", bytes.NewBufferString(reqBody))
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

// TestValidateToken tests the ValidateToken handler
// TestValidateToken 測試 ValidateToken 處理器
func TestValidateToken(t *testing.T) {
	handler, mockService := setupTestHandler()
	router := setupTestRouter(handler)
	router.POST("/validate", handler.ValidateToken)

	t.Run("Valid Token", func(t *testing.T) {
		// Mock service response
		// 模擬服務響應
		mockUserInfo := &UserInfo{
			ID:          1,
			Email:       "test@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
		}
		mockService.On("ValidateToken", "valid-token").Return(mockUserInfo, nil).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("POST", "/validate", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
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
		assert.Contains(t, response.Message, "Token is valid")

		mockService.AssertExpectations(t)
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		// Create request without Authorization header
		// 創建沒有 Authorization 標頭的請求
		req, _ := http.NewRequest("POST", "/validate", nil)
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
		assert.Equal(t, "MISSING_TOKEN", response.Error.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		// Mock service response for invalid token
		// 模擬無效令牌的服務響應
		mockService.On("ValidateToken", "invalid-token").Return(nil, assert.AnError).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("POST", "/validate", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		// Perform request
		// 執行請求
		router.ServeHTTP(w, req)

		// Assert response
		// 斷言響應
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response utils.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)

		mockService.AssertExpectations(t)
	})
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
		mockUserInfo := &UserInfo{
			ID:          1,
			Email:       "test@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
		}
		mockService.On("ValidateToken", "valid-token").Return(mockUserInfo, nil).Once()

		// Create request
		// 創建請求
		req, _ := http.NewRequest("GET", "/profile", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
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
}
