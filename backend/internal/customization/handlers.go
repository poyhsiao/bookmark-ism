package customization

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomizationService interface for dependency injection
type CustomizationService interface {
	CreateTheme(ctx context.Context, userID string, req *CreateThemeRequest) (*ThemeResponse, error)
	GetTheme(ctx context.Context, userID string, themeID uint) (*ThemeResponse, error)
	UpdateTheme(ctx context.Context, userID string, themeID uint, req *UpdateThemeRequest) (*ThemeResponse, error)
	DeleteTheme(ctx context.Context, userID string, themeID uint) error
	ListThemes(ctx context.Context, req *ThemeListRequest) ([]ThemeResponse, int64, error)
	GetUserPreferences(ctx context.Context, userID string) (*UserPreferencesResponse, error)
	UpdateUserPreferences(ctx context.Context, userID string, req *UpdateUserPreferencesRequest) (*UserPreferencesResponse, error)
	GetUserTheme(ctx context.Context, userID string) (*UserThemeResponse, error)
	SetUserTheme(ctx context.Context, userID string, req *SetUserThemeRequest) (*UserThemeResponse, error)
	RateTheme(ctx context.Context, userID string, themeID uint, req *RateThemeRequest) (*ThemeRating, error)
}

// Handler handles HTTP requests for customization features
type Handler struct {
	service CustomizationService
}

// NewHandler creates a new customization handler
func NewHandler(service CustomizationService) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateTheme handles POST /api/v1/customization/themes
func (h *Handler) CreateTheme(c *gin.Context) {
	var req CreateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	theme, err := h.service.CreateTheme(c.Request.Context(), userID.(string), &req)
	if err != nil {
		switch err {
		case ErrInvalidThemeName, ErrInvalidDisplayName, ErrInvalidDescription, ErrInvalidThemeConfig:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		case ErrThemeAlreadyExists:
			c.JSON(http.StatusConflict, NewErrorResponse(err, CodeAlreadyExists, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to create theme"))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"theme": theme})
}

// GetTheme handles GET /api/v1/customization/themes/:id
func (h *Handler) GetTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidThemeID, CodeValidationError, "Invalid theme ID"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	theme, err := h.service.GetTheme(c.Request.Context(), userIDStr, uint(themeID))
	if err != nil {
		switch err {
		case ErrThemeNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		case ErrUnauthorizedTheme:
			c.JSON(http.StatusForbidden, NewErrorResponse(err, CodePermissionDenied, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get theme"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"theme": theme})
}

// UpdateTheme handles PUT /api/v1/customization/themes/:id
func (h *Handler) UpdateTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidThemeID, CodeValidationError, "Invalid theme ID"))
		return
	}

	var req UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	theme, err := h.service.UpdateTheme(c.Request.Context(), userID.(string), uint(themeID), &req)
	if err != nil {
		switch err {
		case ErrThemeNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		case ErrUnauthorizedTheme:
			c.JSON(http.StatusForbidden, NewErrorResponse(err, CodePermissionDenied, err.Error()))
		case ErrInvalidDisplayName, ErrInvalidDescription, ErrInvalidThemeConfig:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to update theme"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"theme": theme})
}

// DeleteTheme handles DELETE /api/v1/customization/themes/:id
func (h *Handler) DeleteTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidThemeID, CodeValidationError, "Invalid theme ID"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	err = h.service.DeleteTheme(c.Request.Context(), userID.(string), uint(themeID))
	if err != nil {
		switch err {
		case ErrThemeNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		case ErrUnauthorizedTheme:
			c.JSON(http.StatusForbidden, NewErrorResponse(err, CodePermissionDenied, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to delete theme"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme deleted successfully"})
}

// ListThemes handles GET /api/v1/customization/themes
func (h *Handler) ListThemes(c *gin.Context) {
	req := ThemeListRequest{
		Page:       1,
		Limit:      20,
		SortBy:     "created_at",
		SortOrder:  "desc",
		PublicOnly: false,
	}

	// Parse query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if search := c.Query("search"); search != "" {
		req.Search = search
	}

	if category := c.Query("category"); category != "" {
		req.Category = category
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	if publicOnlyStr := c.Query("public_only"); publicOnlyStr == "true" {
		req.PublicOnly = true
	}

	themes, total, err := h.service.ListThemes(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to list themes"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"themes": themes,
		"total":  total,
		"page":   req.Page,
		"limit":  req.Limit,
	})
}

// GetUserPreferences handles GET /api/v1/customization/preferences
func (h *Handler) GetUserPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	prefs, err := h.service.GetUserPreferences(c.Request.Context(), userID.(string))
	if err != nil {
		switch err {
		case ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get user preferences"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"preferences": prefs})
}

// UpdateUserPreferences handles PUT /api/v1/customization/preferences
func (h *Handler) UpdateUserPreferences(c *gin.Context) {
	var req UpdateUserPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	prefs, err := h.service.UpdateUserPreferences(c.Request.Context(), userID.(string), &req)
	if err != nil {
		switch err {
		case ErrInvalidUserID, ErrInvalidLanguage, ErrInvalidGridSize, ErrInvalidViewMode,
			ErrInvalidSortBy, ErrInvalidSortOrder, ErrInvalidSyncInterval, ErrInvalidSidebarWidth:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to update user preferences"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"preferences": prefs})
}

// GetUserTheme handles GET /api/v1/customization/theme
func (h *Handler) GetUserTheme(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	userTheme, err := h.service.GetUserTheme(c.Request.Context(), userID.(string))
	if err != nil {
		switch err {
		case ErrThemeNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		case ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get user theme"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_theme": userTheme})
}

// SetUserTheme handles POST /api/v1/customization/theme
func (h *Handler) SetUserTheme(c *gin.Context) {
	var req SetUserThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	userTheme, err := h.service.SetUserTheme(c.Request.Context(), userID.(string), &req)
	if err != nil {
		switch err {
		case ErrThemeNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		case ErrThemeNotPublic:
			c.JSON(http.StatusForbidden, NewErrorResponse(err, CodePermissionDenied, err.Error()))
		case ErrInvalidUserID, ErrInvalidThemeConfig:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to set user theme"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_theme": userTheme})
}

// RateTheme handles POST /api/v1/customization/themes/:id/rate
func (h *Handler) RateTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidThemeID, CodeValidationError, "Invalid theme ID"))
		return
	}

	var req RateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrInvalidUserID, CodeUnauthorized, "User not authenticated"))
		return
	}

	rating, err := h.service.RateTheme(c.Request.Context(), userID.(string), uint(themeID), &req)
	if err != nil {
		switch err {
		case ErrThemeNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		case ErrAlreadyRated:
			c.JSON(http.StatusConflict, NewErrorResponse(err, CodeAlreadyExists, err.Error()))
		case ErrInvalidUserID, ErrInvalidRating, ErrInvalidComment:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to rate theme"))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"rating": rating})
}

// RegisterRoutes registers all customization routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	customization := router.Group("/customization")
	{
		// Theme management
		themes := customization.Group("/themes")
		{
			themes.POST("", h.CreateTheme)
			themes.GET("", h.ListThemes)
			themes.GET("/:id", h.GetTheme)
			themes.PUT("/:id", h.UpdateTheme)
			themes.DELETE("/:id", h.DeleteTheme)
			themes.POST("/:id/rate", h.RateTheme)
		}

		// User preferences
		customization.GET("/preferences", h.GetUserPreferences)
		customization.PUT("/preferences", h.UpdateUserPreferences)

		// User theme
		customization.GET("/theme", h.GetUserTheme)
		customization.POST("/theme", h.SetUserTheme)
	}
}
