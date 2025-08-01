package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a test router with auth middleware
// setupTestRouter 創建帶有認證中間件的測試路由器
func setupTestRouter(jwtConfig *config.JWTConfig, optional bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	if optional {
		router.Use(OptionalAuthMiddleware(jwtConfig))
	} else {
		router.Use(AuthMiddleware(jwtConfig))
	}

	router.GET("/test", func(c *gin.Context) {
		userID := GetUserID(c)
		email := GetUserEmail(c)
		supabaseID := GetSupabaseID(c)
		isAuth := RequireAuth(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":     userID,
			"email":       email,
			"supabase_id": supabaseID,
			"is_auth":     isAuth,
		})
	})

	return router
}

// generateTestToken creates a test JWT token
// generateTestToken 創建測試 JWT 令牌
func generateTestToken(secret string, userID uint, email, username, supabaseID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     float64(userID),
		"email":       email,
		"username":    username,
		"sub":         supabaseID,
		"supabase_id": supabaseID,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour).Unix(),
		"type":        "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// TestAuthMiddleware tests the authentication middleware
// TestAuthMiddleware 測試認證中間件
func TestAuthMiddleware(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:     "test-secret",
		ExpiryHour: 24,
	}

	t.Run("Valid Token", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, false)

		// Generate valid token
		// 生成有效令牌
		token, err := generateTestToken(jwtConfig.Secret, 1, "test@example.com", "testuser", "test-supabase-id")
		require.NoError(t, err)

		// Create request with valid token
		// 創建帶有有效令牌的請求
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check response contains user info
		// 檢查響應包含用戶信息
		assert.Contains(t, w.Body.String(), `"user_id":"1"`)
		assert.Contains(t, w.Body.String(), `"email":"test@example.com"`)
		assert.Contains(t, w.Body.String(), `"supabase_id":"test-supabase-id"`)
		assert.Contains(t, w.Body.String(), `"is_auth":true`)
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, false)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	t.Run("Invalid Authorization Header Format", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, false)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("Invalid Token", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, false)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("Expired Token", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, false)

		// Generate expired token
		// 生成過期令牌
		claims := jwt.MapClaims{
			"user_id": float64(1),
			"email":   "test@example.com",
			"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("Wrong Signing Method", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, false)

		// Create a malformed token
		// 創建一個格式錯誤的令牌
		tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})
}

// TestOptionalAuthMiddleware tests the optional authentication middleware
// TestOptionalAuthMiddleware 測試可選認證中間件
func TestOptionalAuthMiddleware(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:     "test-secret",
		ExpiryHour: 24,
	}

	t.Run("Valid Token", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, true)

		// Generate valid token
		// 生成有效令牌
		token, err := generateTestToken(jwtConfig.Secret, 1, "test@example.com", "testuser", "test-supabase-id")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"user_id":"1"`)
		assert.Contains(t, w.Body.String(), `"is_auth":true`)
	})

	t.Run("No Authorization Header", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, true)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still pass through
		// 應該仍然通過
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"user_id":""`)
		assert.Contains(t, w.Body.String(), `"is_auth":false`)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, true)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still pass through without user info
		// 應該仍然通過但沒有用戶信息
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"user_id":""`)
		assert.Contains(t, w.Body.String(), `"is_auth":false`)
	})

	t.Run("Invalid Authorization Header Format", func(t *testing.T) {
		router := setupTestRouter(jwtConfig, true)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still pass through
		// 應該仍然通過
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"user_id":""`)
		assert.Contains(t, w.Body.String(), `"is_auth":false`)
	})
}

// TestHelperFunctions tests the middleware helper functions
// TestHelperFunctions 測試中間件輔助函數
func TestHelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetUserID", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", "123")

		userID := GetUserID(c)
		assert.Equal(t, "123", userID)
	})

	t.Run("GetUserEmail", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("email", "test@example.com")

		email := GetUserEmail(c)
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("GetSupabaseID", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("supabase_id", "supabase-123")

		supabaseID := GetSupabaseID(c)
		assert.Equal(t, "supabase-123", supabaseID)
	})

	t.Run("RequireAuth - Authenticated", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", "123")

		isAuth := RequireAuth(c)
		assert.True(t, isAuth)
	})

	t.Run("RequireAuth - Not Authenticated", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		isAuth := RequireAuth(c)
		assert.False(t, isAuth)
	})

	t.Run("Empty Values", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		assert.Equal(t, "", GetUserID(c))
		assert.Equal(t, "", GetUserEmail(c))
		assert.Equal(t, "", GetSupabaseID(c))
		assert.False(t, RequireAuth(c))
	})
}
