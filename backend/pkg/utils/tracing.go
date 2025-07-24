// Package utils provides utility functions and middleware for the application
package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceContext holds tracing information for a request
// 保存請求的追蹤信息
type TraceContext struct {
	RequestID string
	UserID    string
	StartTime time.Time
	Logger    *zap.Logger
}

// TracingMiddleware creates a middleware that adds request tracing and structured logging
// 創建添加請求追蹤和結構化日誌記錄的中間件
func TracingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or extract request ID
		// 生成或提取請求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Create trace context
		// 創建追蹤上下文
		trace := &TraceContext{
			RequestID: requestID,
			StartTime: time.Now(),
			Logger:    logger.With(zap.String("request_id", requestID)),
		}

		// Set request ID in response header
		// 在響應標頭中設置請求 ID
		c.Header("X-Request-ID", requestID)

		// Store trace context in Gin context
		// 在 Gin 上下文中存儲追蹤上下文
		c.Set("trace", trace)
		c.Set("request_id", requestID)

		// Log request start
		// 記錄請求開始
		trace.Logger.Info("Request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Process request
		// 處理請求
		c.Next()

		// Calculate request duration
		// 計算請求持續時間
		duration := time.Since(trace.StartTime)

		// Log request completion
		// 記錄請求完成
		trace.Logger.Info("Request completed",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}

// GetTraceFromContext extracts the trace context from Gin context
// 從 Gin 上下文中提取追蹤上下文
func GetTraceFromContext(c *gin.Context) *TraceContext {
	if trace, exists := c.Get("trace"); exists {
		if traceCtx, ok := trace.(*TraceContext); ok {
			return traceCtx
		}
	}
	return nil
}

// LogInfo logs an info message with trace context
// 使用追蹤上下文記錄信息消息
func (t *TraceContext) LogInfo(msg string, fields ...zap.Field) {
	t.Logger.Info(msg, fields...)
}

// LogError logs an error message with trace context
// 使用追蹤上下文記錄錯誤消息
func (t *TraceContext) LogError(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	t.Logger.Error(msg, allFields...)
}

// LogWarn logs a warning message with trace context
// 使用追蹤上下文記錄警告消息
func (t *TraceContext) LogWarn(msg string, fields ...zap.Field) {
	t.Logger.Warn(msg, fields...)
}

// LogDebug logs a debug message with trace context
// 使用追蹤上下文記錄調試消息
func (t *TraceContext) LogDebug(msg string, fields ...zap.Field) {
	t.Logger.Debug(msg, fields...)
}

// WithContext adds the trace context to a Go context
// 將追蹤上下文添加到 Go 上下文中
func (t *TraceContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "trace", t)
}

// GetTraceFromGoContext extracts trace context from Go context
// 從 Go 上下文中提取追蹤上下文
func GetTraceFromGoContext(ctx context.Context) *TraceContext {
	if trace, ok := ctx.Value("trace").(*TraceContext); ok {
		return trace
	}
	return nil
}
