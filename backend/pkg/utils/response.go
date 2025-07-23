package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	requestID := c.GetString("request_id")

	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	}

	c.JSON(http.StatusOK, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	requestID := c.GetString("request_id")

	response := APIResponse{
		Success:   false,
		RequestID: requestID,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	c.JSON(statusCode, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, errors map[string]interface{}) {
	ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", errors)
}

// NotFoundResponse sends a not found error response
func NotFoundResponse(c *gin.Context, resource string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", resource+" not found", nil)
}

// UnauthorizedResponse sends an unauthorized error response
func UnauthorizedResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// InternalErrorResponse sends an internal server error response
func InternalErrorResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}
