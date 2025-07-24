package auth

import (
	"net/http"
	"strconv"

	"bookmark-sync-service/backend/pkg/middleware"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new authentication handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Get trace context for logging
	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("User registration request", zap.String("email", req.Email), zap.String("username", req.Username))
	}

	response, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if trace != nil {
			trace.LogError("Registration failed", err, zap.String("email", req.Email))
		}

		// Check for specific error types
		if err.Error() == "user with email or username already exists" {
			utils.ErrorResponse(c, http.StatusConflict, "USER_EXISTS", "User with this email or username already exists", nil)
			return
		}

		utils.InternalErrorResponse(c, "Registration failed")
		return
	}

	if trace != nil {
		trace.LogInfo("User registered successfully", zap.Uint("user_id", response.User.ID))
	}

	utils.SuccessResponse(c, response, "User registered successfully")
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Get trace context for logging
	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("User login request", zap.String("email", req.Email))
	}

	response, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if trace != nil {
			trace.LogError("Login failed", err, zap.String("email", req.Email))
		}

		// Check for specific error types
		if err.Error() == "invalid credentials" || err.Error() == "user not found" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}

		utils.InternalErrorResponse(c, "Login failed")
		return
	}

	if trace != nil {
		trace.LogInfo("User logged in successfully", zap.Uint("user_id", response.User.ID))
	}

	utils.SuccessResponse(c, response, "Login successful")
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Get trace context for logging
	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Token refresh request")
	}

	response, err := h.service.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		if trace != nil {
			trace.LogError("Token refresh failed", err)
		}

		utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid or expired refresh token", nil)
		return
	}

	if trace != nil {
		trace.LogInfo("Token refreshed successfully", zap.Uint("user_id", response.User.ID))
	}

	utils.SuccessResponse(c, response, "Token refreshed successfully")
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	// Get user ID from middleware
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	// Get trace context for logging
	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("User logout request", zap.Uint("user_id", uint(userID)))
	}

	if err := h.service.Logout(c.Request.Context(), uint(userID)); err != nil {
		if trace != nil {
			trace.LogError("Logout failed", err, zap.Uint("user_id", uint(userID)))
		}

		utils.InternalErrorResponse(c, "Logout failed")
		return
	}

	if trace != nil {
		trace.LogInfo("User logged out successfully", zap.Uint("user_id", uint(userID)))
	}

	utils.SuccessResponse(c, nil, "Logout successful")
}

// ResetPassword handles password reset requests
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Get trace context for logging
	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Password reset request", zap.String("email", req.Email))
	}

	if err := h.service.ResetPassword(c.Request.Context(), &req); err != nil {
		if trace != nil {
			trace.LogError("Password reset failed", err, zap.String("email", req.Email))
		}

		utils.InternalErrorResponse(c, "Password reset failed")
		return
	}

	if trace != nil {
		trace.LogInfo("Password reset email sent", zap.String("email", req.Email))
	}

	utils.SuccessResponse(c, nil, "Password reset email sent")
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from middleware
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", nil)
		return
	}

	// Get trace context for logging
	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Get profile request", zap.Uint("user_id", uint(userID)))
	}

	// For now, we'll use the ValidateToken method to get user info
	// In a real implementation, you might want a separate GetProfile method
	authHeader := c.GetHeader("Authorization")
	tokenString := authHeader[7:] // Remove "Bearer " prefix

	userInfo, err := h.service.ValidateToken(tokenString)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to get profile", err, zap.Uint("user_id", uint(userID)))
		}

		utils.UnauthorizedResponse(c, "Invalid token")
		return
	}

	if trace != nil {
		trace.LogInfo("Profile retrieved successfully", zap.Uint("user_id", uint(userID)))
	}

	utils.SuccessResponse(c, userInfo, "Profile retrieved successfully")
}

// ValidateToken validates a token and returns user info (for internal use)
func (h *Handler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_TOKEN", "Authorization header is required", nil)
		return
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix

	userInfo, err := h.service.ValidateToken(tokenString)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid token")
		return
	}

	utils.SuccessResponse(c, userInfo, "Token is valid")
}
