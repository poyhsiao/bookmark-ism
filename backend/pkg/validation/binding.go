package validation

import (
	"fmt"
	"strconv"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/middleware"
	"bookmark-sync-service/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RequestValidator provides common validation and binding functionality
type RequestValidator struct{}

// NewRequestValidator creates a new request validator
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{}
}

// UserIDFromContext extracts and validates user ID from context
func (v *RequestValidator) UserIDFromContext(c *gin.Context) (uint, error) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		return 0, fmt.Errorf(config.ErrUserNotAuthenticated)
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf(config.ErrInvalidUserID)
	}

	return uint(userID), nil
}

// UserIDFromHeader extracts and validates user ID from header
func (v *RequestValidator) UserIDFromHeader(c *gin.Context) (uint, error) {
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		return 0, fmt.Errorf(config.ErrUserNotAuthenticated)
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf(config.ErrInvalidUserID)
	}

	return uint(userID), nil
}

// IDFromParam extracts and validates ID from URL parameter
func (v *RequestValidator) IDFromParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", paramName)
	}

	return uint(id), nil
}

// BindAndValidateJSON binds JSON request body and validates it
func (v *RequestValidator) BindAndValidateJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("%s: %w", config.ErrInvalidData, err)
	}
	return nil
}

// BindAndValidateQuery binds query parameters and validates them
func (v *RequestValidator) BindAndValidateQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return fmt.Errorf("%s: %w", config.ErrInvalidData, err)
	}
	return nil
}

// BindAndValidateURI binds URI parameters and validates them
func (v *RequestValidator) BindAndValidateURI(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		return fmt.Errorf("%s: %w", config.ErrInvalidData, err)
	}
	return nil
}

// HandleValidationError handles validation errors consistently
func (v *RequestValidator) HandleValidationError(c *gin.Context, err error) {
	utils.ValidationErrorResponse(c, map[string]interface{}{
		"validation_errors": err.Error(),
	})
}

// HandleUnauthorizedError handles unauthorized errors consistently
func (v *RequestValidator) HandleUnauthorizedError(c *gin.Context, message string) {
	utils.UnauthorizedResponse(c, message)
}

// HandleNotFoundError handles not found errors consistently
func (v *RequestValidator) HandleNotFoundError(c *gin.Context, resource string) {
	utils.NotFoundResponse(c, resource)
}

// HandleInternalError handles internal errors consistently
func (v *RequestValidator) HandleInternalError(c *gin.Context, message string) {
	utils.InternalErrorResponse(c, message)
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
	Offset   int `form:"-"`
}

// CalculateOffset calculates the offset for pagination
func (p *PaginationParams) CalculateOffset() {
	p.Offset = (p.Page - 1) * p.PageSize
}

// ValidatePagination validates and binds pagination parameters
func (v *RequestValidator) ValidatePagination(c *gin.Context) (*PaginationParams, error) {
	var params PaginationParams
	if err := v.BindAndValidateQuery(c, &params); err != nil {
		return nil, err
	}

	params.CalculateOffset()
	return &params, nil
}
