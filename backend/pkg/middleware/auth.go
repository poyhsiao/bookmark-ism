package middleware

import (
	"strings"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid token")
			c.Abort()
			return
		}

		if !token.Valid {
			utils.UnauthorizedResponse(c, "Token is not valid")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Set user information in context
			if userID, exists := claims["user_id"]; exists {
				c.Set("user_id", userID)
			}
			if email, exists := claims["email"]; exists {
				c.Set("email", email)
			}
			if supabaseID, exists := claims["sub"]; exists {
				c.Set("supabase_id", supabaseID)
			}
		}

		c.Next()
	}
}

// OptionalAuthMiddleware creates an optional JWT authentication middleware
// This middleware will extract user information if a valid token is provided,
// but won't block the request if no token is provided
func OptionalAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		// Extract claims if token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Set user information in context
			if userID, exists := claims["user_id"]; exists {
				c.Set("user_id", userID)
			}
			if email, exists := claims["email"]; exists {
				c.Set("email", email)
			}
			if supabaseID, exists := claims["sub"]; exists {
				c.Set("supabase_id", supabaseID)
			}
		}

		c.Next()
	}
}

// RequireAuth is a helper function to check if user is authenticated
func RequireAuth(c *gin.Context) bool {
	userID := c.GetString("user_id")
	return userID != ""
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

// GetSupabaseID extracts Supabase user ID from context
func GetSupabaseID(c *gin.Context) string {
	return c.GetString("supabase_id")
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) string {
	return c.GetString("email")
}
