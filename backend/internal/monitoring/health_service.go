package monitoring

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status       HealthStatus `json:"status"`
	ResponseTime string       `json:"response_time,omitempty"`
	Error        string       `json:"error,omitempty"`
	LastCheck    time.Time    `json:"last_check"`
}

// OverallHealth 整体健康状态
type OverallHealth struct {
	Status    HealthStatus             `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
	Version   string                   `json:"version,omitempty"`
	Uptime    string                   `json:"uptime,omitempty"`
}

// HealthService 健康检查服务
type HealthService struct {
	logger    *zap.Logger
	db        *gorm.DB
	rdb       *redis.Client
	startTime time.Time
	version   string

	// 健康检查配置
	timeout time.Duration
}

// NewHealthService 创建健康检查服务
func NewHealthService(logger *zap.Logger, db *gorm.DB, rdb *redis.Client, version string) *HealthService {
	return &HealthService{
		logger:    logger,
		db:        db,
		rdb:       rdb,
		startTime: time.Now(),
		version:   version,
		timeout:   5 * time.Second,
	}
}

// CheckHealth 检查整体健康状态
func (hs *HealthService) CheckHealth(ctx context.Context) *OverallHealth {
	health := &OverallHealth{
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceHealth),
		Version:   hs.version,
		Uptime:    time.Since(hs.startTime).String(),
	}

	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, hs.timeout)
	defer cancel()

	// 检查各个服务
	health.Services["database"] = hs.checkDatabase(checkCtx)
	health.Services["redis"] = hs.checkRedis(checkCtx)
	health.Services["storage"] = hs.checkStorage(checkCtx)

	// 确定整体状态
	health.Status = hs.determineOverallStatus(health.Services)

	return health
}

// checkDatabase 检查数据库健康状态
func (hs *HealthService) checkDatabase(ctx context.Context) ServiceHealth {
	start := time.Now()
	health := ServiceHealth{
		LastCheck: start,
	}

	if hs.db == nil {
		health.Status = HealthStatusUnhealthy
		health.Error = "数据库连接未初始化"
		return health
	}

	// 检查数据库连接
	sqlDB, err := hs.db.DB()
	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Error = "无法获取数据库连接: " + err.Error()
		return health
	}

	// Ping数据库
	err = sqlDB.PingContext(ctx)
	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Error = "数据库ping失败: " + err.Error()
		return health
	}

	// 检查连接池状态
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		health.Status = HealthStatusDegraded
		health.Error = "没有打开的数据库连接"
	} else {
		health.Status = HealthStatusHealthy
	}

	health.ResponseTime = time.Since(start).String()
	return health
}

// checkRedis 检查Redis健康状态
func (hs *HealthService) checkRedis(ctx context.Context) ServiceHealth {
	start := time.Now()
	health := ServiceHealth{
		LastCheck: start,
	}

	if hs.rdb == nil {
		health.Status = HealthStatusUnhealthy
		health.Error = "Redis连接未初始化"
		return health
	}

	// Ping Redis
	result := hs.rdb.Ping(ctx)
	if result.Err() != nil {
		health.Status = HealthStatusUnhealthy
		health.Error = "Redis ping失败: " + result.Err().Error()
		return health
	}

	// 检查连接池状态
	stats := hs.rdb.PoolStats()
	if stats.TotalConns == 0 {
		health.Status = HealthStatusDegraded
		health.Error = "没有可用的Redis连接"
	} else {
		health.Status = HealthStatusHealthy
	}

	health.ResponseTime = time.Since(start).String()
	return health
}

// checkStorage 检查存储服务健康状态
func (hs *HealthService) checkStorage(ctx context.Context) ServiceHealth {
	start := time.Now()
	health := ServiceHealth{
		LastCheck: start,
		Status:    HealthStatusHealthy, // 假设存储服务健康，实际应检查MinIO
		ResponseTime: time.Since(start).String(),
	}

	// TODO: 实现MinIO健康检查
	// 这里可以添加MinIO的健康检查逻辑

	return health
}

// determineOverallStatus 确定整体健康状态
func (hs *HealthService) determineOverallStatus(services map[string]ServiceHealth) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, service := range services {
		switch service.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// HealthHandler 健康检查HTTP处理器
func (hs *HealthService) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		health := hs.CheckHealth(c.Request.Context())

		// 根据健康状态设置HTTP状态码
		var statusCode int
		switch health.Status {
		case HealthStatusHealthy:
			statusCode = http.StatusOK
		case HealthStatusDegraded:
			statusCode = http.StatusOK // 降级但仍可用
		case HealthStatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, health)
	}
}

// ReadinessHandler 就绪检查处理器
func (hs *HealthService) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 就绪检查只检查关键服务
		readiness := map[string]interface{}{
			"ready":     true,
			"timestamp": time.Now(),
		}

		// 检查数据库
		if hs.db != nil {
			sqlDB, err := hs.db.DB()
			if err == nil {
				if err := sqlDB.Ping(); err != nil {
					readiness["ready"] = false
					readiness["error"] = "数据库不可用"
					c.JSON(http.StatusServiceUnavailable, readiness)
					return
				}
			}
		}

		c.JSON(http.StatusOK, readiness)
	}
}

// LivenessHandler 存活检查处理器
func (hs *HealthService) LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 存活检查只检查应用本身是否运行
		liveness := map[string]interface{}{
			"alive":     true,
			"timestamp": time.Now(),
			"uptime":    time.Since(hs.startTime).String(),
		}

		c.JSON(http.StatusOK, liveness)
	}
}

// DetailedHealthHandler 详细健康检查处理器
func (hs *HealthService) DetailedHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		health := hs.CheckHealth(c.Request.Context())

		// 添加额外的详细信息
		detailed := map[string]interface{}{
			"health": health,
			"system": hs.getSystemInfo(),
		}

		var statusCode int
		switch health.Status {
		case HealthStatusHealthy:
			statusCode = http.StatusOK
		case HealthStatusDegraded:
			statusCode = http.StatusOK
		case HealthStatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, detailed)
	}
}

// getSystemInfo 获取系统信息
func (hs *HealthService) getSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"version":    hs.version,
		"start_time": hs.startTime,
		"uptime":     time.Since(hs.startTime).String(),
	}

	// 添加数据库连接池信息
	if hs.db != nil {
		if sqlDB, err := hs.db.DB(); err == nil {
			stats := sqlDB.Stats()
			info["database_pool"] = map[string]interface{}{
				"open_connections": stats.OpenConnections,
				"in_use":          stats.InUse,
				"idle":            stats.Idle,
			}
		}
	}

	// 添加Redis连接池信息
	if hs.rdb != nil {
		stats := hs.rdb.PoolStats()
		info["redis_pool"] = map[string]interface{}{
			"total_conns": stats.TotalConns,
			"idle_conns":  stats.IdleConns,
			"stale_conns": stats.StaleConns,
		}
	}

	return info
}

// RegisterHealthRoutes 注册健康检查路由
func (hs *HealthService) RegisterHealthRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("", hs.HealthHandler())
		health.GET("/live", hs.LivenessHandler())
		health.GET("/ready", hs.ReadinessHandler())
		health.GET("/detailed", hs.DetailedHealthHandler())
	}
}