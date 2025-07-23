package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceContext represents tracing information
type TraceContext struct {
	RequestID string
	UserID    string
	StartTime time.Time
	Logger    *zap.Logger
}

// NewTraceContext creates a new trace context
func NewTraceContext(c *gin.Context, logger *zap.Logger) *TraceContext {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = uuid.New().String()
		c.Set("request_id", requestID)
	}

	userID := c.GetString("user_id")

	return &TraceContext{
		RequestID: requestID,
		UserID:    userID,
		StartTime: time.Now(),
		Logger:    logger.With(zap.String("request_id", requestID)),
	}
}

// WithContext adds tracing information to context
func (tc *TraceContext) WithContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "request_id", tc.RequestID)
	ctx = context.WithValue(ctx, "user_id", tc.UserID)
	return ctx
}

// LogInfo logs an info message with trace context
func (tc *TraceContext) LogInfo(message string, fields ...zap.Field) {
	allFields := append(fields, zap.String("user_id", tc.UserID))
	tc.Logger.Info(message, allFields...)
}

// LogError logs an error message with trace context
func (tc *TraceContext) LogError(message string, err error, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("user_id", tc.UserID),
		zap.Error(err),
	)
	tc.Logger.Error(message, allFields...)
}

// LogWarn logs a warning message with trace context
func (tc *TraceContext) LogWarn(message string, fields ...zap.Field) {
	allFields := append(fields, zap.String("user_id", tc.UserID))
	tc.Logger.Warn(message, allFields...)
}

// Duration returns the duration since the trace started
func (tc *TraceContext) Duration() time.Duration {
	return time.Since(tc.StartTime)
}

// TracingMiddleware creates a middleware for request tracing
func TracingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		trace := NewTraceContext(c, logger)
		c.Set("trace", trace)

		// Add request ID to response header
		c.Header("X-Request-ID", trace.RequestID)

		c.Next()

		// Log request completion
		trace.LogInfo("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", trace.Duration()),
		)
	}
}

// GetTraceFromContext extracts trace context from Gin context
func GetTraceFromContext(c *gin.Context) *TraceContext {
	if trace, exists := c.Get("trace"); exists {
		return trace.(*TraceContext)
	}
	return nil
}
