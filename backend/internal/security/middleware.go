package security

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	logger      *zap.Logger
	redis       *redis.Client
	rateLimiter *RateLimiter
	validator   *InputValidator
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(logger *zap.Logger, redis *redis.Client) *SecurityMiddleware {
	return &SecurityMiddleware{
		logger:      logger,
		redis:       redis,
		rateLimiter: NewRateLimiter(redis, logger),
		validator:   NewInputValidator(logger),
	}
}

// SecurityHeaders 安全头中间件
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HTTPS 强制重定向
		if c.GetHeader("X-Forwarded-Proto") == "http" {
			c.Redirect(http.StatusMovedPermanently, "https://"+c.Request.Host+c.Request.RequestURI)
			return
		}

		// 设置安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 移除服务器信息
		c.Header("Server", "")

		c.Next()
	}
}

// CORS CORS中间件
func (sm *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 允许的域名列表
		allowedOrigins := []string{
			"http://localhost:3000",
			"https://localhost:3000",
			"chrome-extension://",
			"moz-extension://",
			"safari-web-extension://",
		}

		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || strings.HasPrefix(origin, allowedOrigin) {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimiting 速率限制中间件
func (sm *SecurityMiddleware) RateLimiting(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		allowed, err := sm.rateLimiter.IsAllowed(c.Request.Context(), clientIP, requests, window)
		if err != nil {
			sm.logger.Error("Rate limiting error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
			c.Abort()
			return
		}

		if !allowed {
			sm.logger.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InputValidation 输入验证中间件
func (sm *SecurityMiddleware) InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证URL路径
		if sm.validator.ContainsSQLInjection(c.Request.URL.Path) {
			sm.logger.Warn("SQL injection detected in URL path",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path))

			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
			c.Abort()
			return
		}

		// 验证查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if sm.validator.ContainsSQLInjection(value) {
					sm.logger.Warn("SQL injection detected in query parameter",
						zap.String("ip", c.ClientIP()),
						zap.String("param", key),
						zap.String("value", value))

					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数"})
					c.Abort()
					return
				}

				if sm.validator.ContainsXSS(value) {
					sm.logger.Warn("XSS detected in query parameter",
						zap.String("ip", c.ClientIP()),
						zap.String("param", key),
						zap.String("value", value))

					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数"})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// RequestSizeLimit 请求大小限制中间件
func (sm *SecurityMiddleware) RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			sm.logger.Warn("Request size exceeded",
				zap.String("ip", c.ClientIP()),
				zap.Int64("content_length", c.Request.ContentLength),
				zap.Int64("max_size", maxSize))

			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "请求体过大",
				"max_size": maxSize,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InputValidator 输入验证器
type InputValidator struct {
	logger           *zap.Logger
	sqlInjectionRe   *regexp.Regexp
	xssRe           *regexp.Regexp
	pathTraversalRe *regexp.Regexp
}

// NewInputValidator 创建输入验证器
func NewInputValidator(logger *zap.Logger) *InputValidator {
	return &InputValidator{
		logger: logger,
		sqlInjectionRe: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|alert|confirm|prompt|\||&|;|'|"|<|>|\+|\-|\*|\/|%|\(|\)|\[|\]|\{|\})`),
		xssRe:          regexp.MustCompile(`(?i)(<script|</script|javascript:|vbscript:|onload=|onerror=|onclick=|onmouseover=|onfocus=|onblur=|onchange=|onsubmit=)`),
		pathTraversalRe: regexp.MustCompile(`(\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e%5c)`),
	}
}

// ContainsSQLInjection 检查是否包含SQL注入
func (iv *InputValidator) ContainsSQLInjection(input string) bool {
	return iv.sqlInjectionRe.MatchString(input)
}

// ContainsXSS 检查是否包含XSS攻击
func (iv *InputValidator) ContainsXSS(input string) bool {
	return iv.xssRe.MatchString(input)
}

// ContainsPathTraversal 检查是否包含路径遍历攻击
func (iv *InputValidator) ContainsPathTraversal(input string) bool {
	return iv.pathTraversalRe.MatchString(input)
}

// SanitizeInput 清理输入
func (iv *InputValidator) SanitizeInput(input string) string {
	// 移除潜在的恶意字符
	cleaned := strings.ReplaceAll(input, "<script", "&lt;script")
	cleaned = strings.ReplaceAll(cleaned, "</script", "&lt;/script")
	cleaned = strings.ReplaceAll(cleaned, "javascript:", "")
	cleaned = strings.ReplaceAll(cleaned, "vbscript:", "")
	cleaned = strings.ReplaceAll(cleaned, "onload=", "")
	cleaned = strings.ReplaceAll(cleaned, "onerror=", "")

	return cleaned
}

// RateLimiter 速率限制器
type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(redis *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		logger: logger,
	}
}

// IsAllowed 检查是否允许请求
func (rl *RateLimiter) IsAllowed(ctx context.Context, clientIP string, maxRequests int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", clientIP)

	// 使用滑动窗口算法
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe := rl.redis.Pipeline()

	// 移除窗口外的记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// 获取当前窗口内的请求数
	pipe.ZCard(ctx, key)

	// 设置过期时间
	pipe.Expire(ctx, key, window)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// 获取请求数
	count := results[2].(*redis.IntCmd).Val()

	return count <= int64(maxRequests), nil
}

// IPBlocker IP封禁器
type IPBlocker struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewIPBlocker 创建IP封禁器
func NewIPBlocker(redis *redis.Client, logger *zap.Logger) *IPBlocker {
	return &IPBlocker{
		redis:  redis,
		logger: logger,
	}
}

// BlockIP 封禁IP
func (ib *IPBlocker) BlockIP(ctx context.Context, ip string, duration time.Duration, reason string) error {
	key := fmt.Sprintf("blocked_ip:%s", ip)

	blockInfo := map[string]interface{}{
		"blocked_at": time.Now().Unix(),
		"reason":     reason,
		"duration":   int64(duration.Seconds()),
	}

	err := ib.redis.HMSet(ctx, key, blockInfo).Err()
	if err != nil {
		return err
	}

	err = ib.redis.Expire(ctx, key, duration).Err()
	if err != nil {
		return err
	}

	ib.logger.Info("IP blocked",
		zap.String("ip", ip),
		zap.String("reason", reason),
		zap.Duration("duration", duration))

	return nil
}

// IsBlocked 检查IP是否被封禁
func (ib *IPBlocker) IsBlocked(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("blocked_ip:%s", ip)
	exists, err := ib.redis.Exists(ctx, key).Result()
	return exists > 0, err
}

// IPBlocking IP封禁中间件
func (sm *SecurityMiddleware) IPBlocking() gin.HandlerFunc {
	blocker := NewIPBlocker(sm.redis, sm.logger)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		blocked, err := blocker.IsBlocked(c.Request.Context(), clientIP)
		if err != nil {
			sm.logger.Error("Failed to check IP block status", zap.Error(err))
			c.Next()
			return
		}

		if blocked {
			sm.logger.Warn("Blocked IP attempted access",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path))

			c.JSON(http.StatusForbidden, gin.H{
				"error": "您的IP地址已被临时封禁",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}