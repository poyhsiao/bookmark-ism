package automation

import "errors"

// Common automation errors
var (
	// Webhook errors
	ErrWebhookEndpointNotFound = errors.New("webhook endpoint not found")
	ErrWebhookEndpointExists   = errors.New("webhook endpoint already exists")
	ErrWebhookDeliveryFailed   = errors.New("webhook delivery failed")
	ErrWebhookInvalidSignature = errors.New("invalid webhook signature")
	ErrWebhookTimeout          = errors.New("webhook request timeout")
	ErrWebhookInvalidURL       = errors.New("invalid webhook URL")
	ErrWebhookInvalidEvent     = errors.New("invalid webhook event")

	// RSS Feed errors
	ErrRSSFeedNotFound         = errors.New("RSS feed not found")
	ErrRSSFeedExists           = errors.New("RSS feed already exists")
	ErrRSSFeedInvalidPublicKey = errors.New("invalid RSS feed public key")
	ErrRSSFeedGenerationFailed = errors.New("RSS feed generation failed")
	ErrRSSFeedInactive         = errors.New("RSS feed is inactive")

	// Bulk Operation errors
	ErrBulkOperationNotFound      = errors.New("bulk operation not found")
	ErrBulkOperationInProgress    = errors.New("bulk operation already in progress")
	ErrBulkOperationCompleted     = errors.New("bulk operation already completed")
	ErrBulkOperationCancelled     = errors.New("bulk operation was cancelled")
	ErrBulkOperationFailed        = errors.New("bulk operation failed")
	ErrBulkOperationInvalidType   = errors.New("invalid bulk operation type")
	ErrBulkOperationInvalidParams = errors.New("invalid bulk operation parameters")

	// Backup Job errors
	ErrBackupJobNotFound    = errors.New("backup job not found")
	ErrBackupJobInProgress  = errors.New("backup job already in progress")
	ErrBackupJobCompleted   = errors.New("backup job already completed")
	ErrBackupJobFailed      = errors.New("backup job failed")
	ErrBackupJobInvalidType = errors.New("invalid backup job type")
	ErrBackupFileNotFound   = errors.New("backup file not found")
	ErrBackupFileCorrupted  = errors.New("backup file is corrupted")

	// API Integration errors
	ErrAPIIntegrationNotFound    = errors.New("API integration not found")
	ErrAPIIntegrationExists      = errors.New("API integration already exists")
	ErrAPIIntegrationInactive    = errors.New("API integration is inactive")
	ErrAPIIntegrationAuthFailed  = errors.New("API integration authentication failed")
	ErrAPIIntegrationRateLimit   = errors.New("API integration rate limit exceeded")
	ErrAPIIntegrationTimeout     = errors.New("API integration request timeout")
	ErrAPIIntegrationInvalidType = errors.New("invalid API integration type")
	ErrAPIIntegrationSyncFailed  = errors.New("API integration sync failed")

	// Automation Rule errors
	ErrAutomationRuleNotFound         = errors.New("automation rule not found")
	ErrAutomationRuleExists           = errors.New("automation rule already exists")
	ErrAutomationRuleInactive         = errors.New("automation rule is inactive")
	ErrAutomationRuleInvalidTrigger   = errors.New("invalid automation rule trigger")
	ErrAutomationRuleInvalidCondition = errors.New("invalid automation rule condition")
	ErrAutomationRuleInvalidAction    = errors.New("invalid automation rule action")
	ErrAutomationRuleExecutionFailed  = errors.New("automation rule execution failed")

	// General errors
	ErrUserNotAuthenticated = errors.New("user not authenticated")
	ErrUserNotAuthorized    = errors.New("user not authorized")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrInvalidParameters    = errors.New("invalid parameters")
	ErrResourceNotFound     = errors.New("resource not found")
	ErrResourceExists       = errors.New("resource already exists")
	ErrInternalServerError  = errors.New("internal server error")
	ErrServiceUnavailable   = errors.New("service unavailable")
	ErrDatabaseError        = errors.New("database error")
	ErrNetworkError         = errors.New("network error")
)

// ErrorCode represents an error code for API responses
type ErrorCode string

const (
	// Webhook error codes
	CodeWebhookEndpointNotFound ErrorCode = "WEBHOOK_ENDPOINT_NOT_FOUND"
	CodeWebhookEndpointExists   ErrorCode = "WEBHOOK_ENDPOINT_EXISTS"
	CodeWebhookDeliveryFailed   ErrorCode = "WEBHOOK_DELIVERY_FAILED"
	CodeWebhookInvalidSignature ErrorCode = "WEBHOOK_INVALID_SIGNATURE"
	CodeWebhookTimeout          ErrorCode = "WEBHOOK_TIMEOUT"
	CodeWebhookInvalidURL       ErrorCode = "WEBHOOK_INVALID_URL"
	CodeWebhookInvalidEvent     ErrorCode = "WEBHOOK_INVALID_EVENT"

	// RSS Feed error codes
	CodeRSSFeedNotFound         ErrorCode = "RSS_FEED_NOT_FOUND"
	CodeRSSFeedExists           ErrorCode = "RSS_FEED_EXISTS"
	CodeRSSFeedInvalidPublicKey ErrorCode = "RSS_FEED_INVALID_PUBLIC_KEY"
	CodeRSSFeedGenerationFailed ErrorCode = "RSS_FEED_GENERATION_FAILED"
	CodeRSSFeedInactive         ErrorCode = "RSS_FEED_INACTIVE"

	// Bulk Operation error codes
	CodeBulkOperationNotFound      ErrorCode = "BULK_OPERATION_NOT_FOUND"
	CodeBulkOperationInProgress    ErrorCode = "BULK_OPERATION_IN_PROGRESS"
	CodeBulkOperationCompleted     ErrorCode = "BULK_OPERATION_COMPLETED"
	CodeBulkOperationCancelled     ErrorCode = "BULK_OPERATION_CANCELLED"
	CodeBulkOperationFailed        ErrorCode = "BULK_OPERATION_FAILED"
	CodeBulkOperationInvalidType   ErrorCode = "BULK_OPERATION_INVALID_TYPE"
	CodeBulkOperationInvalidParams ErrorCode = "BULK_OPERATION_INVALID_PARAMS"

	// Backup Job error codes
	CodeBackupJobNotFound    ErrorCode = "BACKUP_JOB_NOT_FOUND"
	CodeBackupJobInProgress  ErrorCode = "BACKUP_JOB_IN_PROGRESS"
	CodeBackupJobCompleted   ErrorCode = "BACKUP_JOB_COMPLETED"
	CodeBackupJobFailed      ErrorCode = "BACKUP_JOB_FAILED"
	CodeBackupJobInvalidType ErrorCode = "BACKUP_JOB_INVALID_TYPE"
	CodeBackupFileNotFound   ErrorCode = "BACKUP_FILE_NOT_FOUND"
	CodeBackupFileCorrupted  ErrorCode = "BACKUP_FILE_CORRUPTED"

	// API Integration error codes
	CodeAPIIntegrationNotFound    ErrorCode = "API_INTEGRATION_NOT_FOUND"
	CodeAPIIntegrationExists      ErrorCode = "API_INTEGRATION_EXISTS"
	CodeAPIIntegrationInactive    ErrorCode = "API_INTEGRATION_INACTIVE"
	CodeAPIIntegrationAuthFailed  ErrorCode = "API_INTEGRATION_AUTH_FAILED"
	CodeAPIIntegrationRateLimit   ErrorCode = "API_INTEGRATION_RATE_LIMIT"
	CodeAPIIntegrationTimeout     ErrorCode = "API_INTEGRATION_TIMEOUT"
	CodeAPIIntegrationInvalidType ErrorCode = "API_INTEGRATION_INVALID_TYPE"
	CodeAPIIntegrationSyncFailed  ErrorCode = "API_INTEGRATION_SYNC_FAILED"

	// Automation Rule error codes
	CodeAutomationRuleNotFound         ErrorCode = "AUTOMATION_RULE_NOT_FOUND"
	CodeAutomationRuleExists           ErrorCode = "AUTOMATION_RULE_EXISTS"
	CodeAutomationRuleInactive         ErrorCode = "AUTOMATION_RULE_INACTIVE"
	CodeAutomationRuleInvalidTrigger   ErrorCode = "AUTOMATION_RULE_INVALID_TRIGGER"
	CodeAutomationRuleInvalidCondition ErrorCode = "AUTOMATION_RULE_INVALID_CONDITION"
	CodeAutomationRuleInvalidAction    ErrorCode = "AUTOMATION_RULE_INVALID_ACTION"
	CodeAutomationRuleExecutionFailed  ErrorCode = "AUTOMATION_RULE_EXECUTION_FAILED"

	// General error codes
	CodeUserNotAuthenticated ErrorCode = "USER_NOT_AUTHENTICATED"
	CodeUserNotAuthorized    ErrorCode = "USER_NOT_AUTHORIZED"
	CodeInvalidRequest       ErrorCode = "INVALID_REQUEST"
	CodeInvalidParameters    ErrorCode = "INVALID_PARAMETERS"
	CodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	CodeResourceExists       ErrorCode = "RESOURCE_EXISTS"
	CodeInternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	CodeDatabaseError        ErrorCode = "DATABASE_ERROR"
	CodeNetworkError         ErrorCode = "NETWORK_ERROR"
)

// AutomationError represents a structured error for the automation service
type AutomationError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AutomationError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// NewAutomationError creates a new automation error
func NewAutomationError(code ErrorCode, message string, details ...string) *AutomationError {
	err := &AutomationError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// Error mapping functions

// MapWebhookError maps webhook errors to structured errors
func MapWebhookError(err error) *AutomationError {
	switch err {
	case ErrWebhookEndpointNotFound:
		return NewAutomationError(CodeWebhookEndpointNotFound, "Webhook endpoint not found")
	case ErrWebhookEndpointExists:
		return NewAutomationError(CodeWebhookEndpointExists, "Webhook endpoint already exists")
	case ErrWebhookDeliveryFailed:
		return NewAutomationError(CodeWebhookDeliveryFailed, "Webhook delivery failed")
	case ErrWebhookInvalidSignature:
		return NewAutomationError(CodeWebhookInvalidSignature, "Invalid webhook signature")
	case ErrWebhookTimeout:
		return NewAutomationError(CodeWebhookTimeout, "Webhook request timeout")
	case ErrWebhookInvalidURL:
		return NewAutomationError(CodeWebhookInvalidURL, "Invalid webhook URL")
	case ErrWebhookInvalidEvent:
		return NewAutomationError(CodeWebhookInvalidEvent, "Invalid webhook event")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}

// MapRSSFeedError maps RSS feed errors to structured errors
func MapRSSFeedError(err error) *AutomationError {
	switch err {
	case ErrRSSFeedNotFound:
		return NewAutomationError(CodeRSSFeedNotFound, "RSS feed not found")
	case ErrRSSFeedExists:
		return NewAutomationError(CodeRSSFeedExists, "RSS feed already exists")
	case ErrRSSFeedInvalidPublicKey:
		return NewAutomationError(CodeRSSFeedInvalidPublicKey, "Invalid RSS feed public key")
	case ErrRSSFeedGenerationFailed:
		return NewAutomationError(CodeRSSFeedGenerationFailed, "RSS feed generation failed")
	case ErrRSSFeedInactive:
		return NewAutomationError(CodeRSSFeedInactive, "RSS feed is inactive")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}

// MapBulkOperationError maps bulk operation errors to structured errors
func MapBulkOperationError(err error) *AutomationError {
	switch err {
	case ErrBulkOperationNotFound:
		return NewAutomationError(CodeBulkOperationNotFound, "Bulk operation not found")
	case ErrBulkOperationInProgress:
		return NewAutomationError(CodeBulkOperationInProgress, "Bulk operation already in progress")
	case ErrBulkOperationCompleted:
		return NewAutomationError(CodeBulkOperationCompleted, "Bulk operation already completed")
	case ErrBulkOperationCancelled:
		return NewAutomationError(CodeBulkOperationCancelled, "Bulk operation was cancelled")
	case ErrBulkOperationFailed:
		return NewAutomationError(CodeBulkOperationFailed, "Bulk operation failed")
	case ErrBulkOperationInvalidType:
		return NewAutomationError(CodeBulkOperationInvalidType, "Invalid bulk operation type")
	case ErrBulkOperationInvalidParams:
		return NewAutomationError(CodeBulkOperationInvalidParams, "Invalid bulk operation parameters")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}

// MapBackupJobError maps backup job errors to structured errors
func MapBackupJobError(err error) *AutomationError {
	switch err {
	case ErrBackupJobNotFound:
		return NewAutomationError(CodeBackupJobNotFound, "Backup job not found")
	case ErrBackupJobInProgress:
		return NewAutomationError(CodeBackupJobInProgress, "Backup job already in progress")
	case ErrBackupJobCompleted:
		return NewAutomationError(CodeBackupJobCompleted, "Backup job already completed")
	case ErrBackupJobFailed:
		return NewAutomationError(CodeBackupJobFailed, "Backup job failed")
	case ErrBackupJobInvalidType:
		return NewAutomationError(CodeBackupJobInvalidType, "Invalid backup job type")
	case ErrBackupFileNotFound:
		return NewAutomationError(CodeBackupFileNotFound, "Backup file not found")
	case ErrBackupFileCorrupted:
		return NewAutomationError(CodeBackupFileCorrupted, "Backup file is corrupted")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}

// MapAPIIntegrationError maps API integration errors to structured errors
func MapAPIIntegrationError(err error) *AutomationError {
	switch err {
	case ErrAPIIntegrationNotFound:
		return NewAutomationError(CodeAPIIntegrationNotFound, "API integration not found")
	case ErrAPIIntegrationExists:
		return NewAutomationError(CodeAPIIntegrationExists, "API integration already exists")
	case ErrAPIIntegrationInactive:
		return NewAutomationError(CodeAPIIntegrationInactive, "API integration is inactive")
	case ErrAPIIntegrationAuthFailed:
		return NewAutomationError(CodeAPIIntegrationAuthFailed, "API integration authentication failed")
	case ErrAPIIntegrationRateLimit:
		return NewAutomationError(CodeAPIIntegrationRateLimit, "API integration rate limit exceeded")
	case ErrAPIIntegrationTimeout:
		return NewAutomationError(CodeAPIIntegrationTimeout, "API integration request timeout")
	case ErrAPIIntegrationInvalidType:
		return NewAutomationError(CodeAPIIntegrationInvalidType, "Invalid API integration type")
	case ErrAPIIntegrationSyncFailed:
		return NewAutomationError(CodeAPIIntegrationSyncFailed, "API integration sync failed")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}

// MapAutomationRuleError maps automation rule errors to structured errors
func MapAutomationRuleError(err error) *AutomationError {
	switch err {
	case ErrAutomationRuleNotFound:
		return NewAutomationError(CodeAutomationRuleNotFound, "Automation rule not found")
	case ErrAutomationRuleExists:
		return NewAutomationError(CodeAutomationRuleExists, "Automation rule already exists")
	case ErrAutomationRuleInactive:
		return NewAutomationError(CodeAutomationRuleInactive, "Automation rule is inactive")
	case ErrAutomationRuleInvalidTrigger:
		return NewAutomationError(CodeAutomationRuleInvalidTrigger, "Invalid automation rule trigger")
	case ErrAutomationRuleInvalidCondition:
		return NewAutomationError(CodeAutomationRuleInvalidCondition, "Invalid automation rule condition")
	case ErrAutomationRuleInvalidAction:
		return NewAutomationError(CodeAutomationRuleInvalidAction, "Invalid automation rule action")
	case ErrAutomationRuleExecutionFailed:
		return NewAutomationError(CodeAutomationRuleExecutionFailed, "Automation rule execution failed")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}

// MapGeneralError maps general errors to structured errors
func MapGeneralError(err error) *AutomationError {
	switch err {
	case ErrUserNotAuthenticated:
		return NewAutomationError(CodeUserNotAuthenticated, "User not authenticated")
	case ErrUserNotAuthorized:
		return NewAutomationError(CodeUserNotAuthorized, "User not authorized")
	case ErrInvalidRequest:
		return NewAutomationError(CodeInvalidRequest, "Invalid request")
	case ErrInvalidParameters:
		return NewAutomationError(CodeInvalidParameters, "Invalid parameters")
	case ErrResourceNotFound:
		return NewAutomationError(CodeResourceNotFound, "Resource not found")
	case ErrResourceExists:
		return NewAutomationError(CodeResourceExists, "Resource already exists")
	case ErrServiceUnavailable:
		return NewAutomationError(CodeServiceUnavailable, "Service unavailable")
	case ErrDatabaseError:
		return NewAutomationError(CodeDatabaseError, "Database error")
	case ErrNetworkError:
		return NewAutomationError(CodeNetworkError, "Network error")
	default:
		return NewAutomationError(CodeInternalServerError, "Internal server error", err.Error())
	}
}
