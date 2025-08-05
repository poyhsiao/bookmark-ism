package user

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/middleware"
	"bookmark-sync-service/backend/pkg/utils"
	"bookmark-sync-service/backend/pkg/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceInterface defines the interface for user service operations
type ServiceInterface interface {
	GetProfile(ctx context.Context, userID uint) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*UserProfile, error)
	UpdatePreferences(ctx context.Context, userID uint, req *UpdatePreferencesRequest) (*UserProfile, error)
	UploadAvatar(ctx context.Context, userID uint, imageData []byte, contentType string) (*UserProfile, error)
	ExportUserData(ctx context.Context, userID uint) (map[string]interface{}, error)
	DeleteUser(ctx context.Context, userID uint) error
}

// Handler handles HTTP requests for user operations
type Handler struct {
	service   ServiceInterface
	logger    *zap.Logger
	validator *validation.RequestValidator
}

// NewHandler creates a new user handler
func NewHandler(service ServiceInterface, logger *zap.Logger) *Handler {
	return &Handler{
		service:   service,
		logger:    logger,
		validator: validation.NewRequestValidator(),
	}
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, err := h.validator.UserIDFromContext(c)
	if err != nil {
		h.validator.HandleUnauthorizedError(c, config.ErrUserNotAuthenticated)
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Get profile request", zap.Uint("user_id", userID))
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to get profile", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			h.validator.HandleNotFoundError(c, "User")
			return
		}

		h.validator.HandleInternalError(c, "Failed to get profile")
		return
	}

	if trace != nil {
		trace.LogInfo("Profile retrieved successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, profile, "Profile retrieved successfully")
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, err := h.validator.UserIDFromContext(c)
	if err != nil {
		h.validator.HandleUnauthorizedError(c, config.ErrUserNotAuthenticated)
		return
	}

	var req UpdateProfileRequest
	if err := h.validator.BindAndValidateJSON(c, &req); err != nil {
		h.validator.HandleValidationError(c, err)
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Update profile request", zap.Uint("user_id", userID))
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to update profile", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			utils.NotFoundResponse(c, "User")
			return
		}

		if err.Error() == "username already taken" {
			utils.ErrorResponse(c, http.StatusConflict, "USERNAME_TAKEN", "Username is already taken", nil)
			return
		}

		utils.InternalErrorResponse(c, "Failed to update profile")
		return
	}

	if trace != nil {
		trace.LogInfo("Profile updated successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, profile, "Profile updated successfully")
}

// GetPreferences returns the current user's preferences
func (h *Handler) GetPreferences(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Get preferences request", zap.Uint("user_id", userID))
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to get preferences", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			utils.NotFoundResponse(c, "User")
			return
		}

		utils.InternalErrorResponse(c, "Failed to get preferences")
		return
	}

	if trace != nil {
		trace.LogInfo("Preferences retrieved successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, profile.Preferences, "Preferences retrieved successfully")
}

// UpdatePreferences updates the current user's preferences
func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Update preferences request", zap.Uint("user_id", userID))
	}

	profile, err := h.service.UpdatePreferences(c.Request.Context(), userID, &req)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to update preferences", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			utils.NotFoundResponse(c, "User")
			return
		}

		utils.InternalErrorResponse(c, "Failed to update preferences")
		return
	}

	if trace != nil {
		trace.LogInfo("Preferences updated successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, profile.Preferences, "Preferences updated successfully")
}

// UploadAvatar uploads a user's avatar image
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE", "No file uploaded or invalid file", nil)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "Only image files are allowed", nil)
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "FILE_TOO_LARGE", "File size must be less than 5MB", nil)
		return
	}

	// Read file data
	imageData, err := io.ReadAll(file)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to read file")
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Upload avatar request", zap.Uint("user_id", userID), zap.String("content_type", contentType))
	}

	profile, err := h.service.UploadAvatar(c.Request.Context(), userID, imageData, contentType)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to upload avatar", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			utils.NotFoundResponse(c, "User")
			return
		}

		utils.InternalErrorResponse(c, "Failed to upload avatar")
		return
	}

	if trace != nil {
		trace.LogInfo("Avatar uploaded successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, gin.H{
		"avatar_url": profile.Avatar,
		"profile":    profile,
	}, "Avatar uploaded successfully")
}

// GetStats returns the current user's statistics
func (h *Handler) GetStats(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Get stats request", zap.Uint("user_id", userID))
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to get stats", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			utils.NotFoundResponse(c, "User")
			return
		}

		utils.InternalErrorResponse(c, "Failed to get stats")
		return
	}

	if trace != nil {
		trace.LogInfo("Stats retrieved successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, profile.Stats, "Stats retrieved successfully")
}

// ExportData exports all user data for GDPR compliance
func (h *Handler) ExportData(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Export data request", zap.Uint("user_id", userID))
	}

	exportData, err := h.service.ExportUserData(c.Request.Context(), userID)
	if err != nil {
		if trace != nil {
			trace.LogError("Failed to export data", err, zap.Uint("user_id", userID))
		}

		if err.Error() == "user not found" {
			utils.NotFoundResponse(c, "User")
			return
		}

		utils.InternalErrorResponse(c, "Failed to export data")
		return
	}

	if trace != nil {
		trace.LogInfo("Data exported successfully", zap.Uint("user_id", userID))
	}

	// Set headers for file download
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=user_data_export.json")

	utils.SuccessResponse(c, exportData, "Data exported successfully")
}

// DeleteAccount deletes the current user's account
func (h *Handler) DeleteAccount(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Require confirmation
	confirmation := c.Query("confirm")
	if confirmation != "DELETE_MY_ACCOUNT" {
		utils.ErrorResponse(c, http.StatusBadRequest, "CONFIRMATION_REQUIRED",
			"Account deletion requires confirmation. Add ?confirm=DELETE_MY_ACCOUNT to the request", nil)
		return
	}

	trace := utils.GetTraceFromContext(c)
	if trace != nil {
		trace.LogInfo("Delete account request", zap.Uint("user_id", userID))
	}

	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		if trace != nil {
			trace.LogError("Failed to delete account", err, zap.Uint("user_id", userID))
		}

		utils.InternalErrorResponse(c, "Failed to delete account")
		return
	}

	if trace != nil {
		trace.LogInfo("Account deleted successfully", zap.Uint("user_id", userID))
	}

	utils.SuccessResponse(c, nil, "Account deleted successfully")
}

// getUserIDFromContext extracts user ID from the request context
func (h *Handler) getUserIDFromContext(c *gin.Context) (uint, error) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		return 0, fmt.Errorf("user not authenticated")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}

	return uint(userID), nil
}

// Helper function to format error
func (h *Handler) formatError(err error) string {
	return err.Error()
}
