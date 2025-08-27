package monitoring

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MetricsService 监控指标服务
type MetricsService struct {
	logger *zap.Logger
	db     *gorm.DB
	rdb    *redis.Client

	// HTTP指标
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// 数据库指标
	dbConnections *prometheus.GaugeVec
	dbQueryDuration *prometheus.HistogramVec

	// Redis指标
	redisConnections *prometheus.GaugeVec
	redisCommandDuration *prometheus.HistogramVec

	// 应用指标
	activeConnections prometheus.Gauge
	businessMetrics   *prometheus.GaugeVec

	// Go运行时指标（已由prometheus自动提供）
}

// NewMetricsService 创建新的监控服务
func NewMetricsService(logger *zap.Logger, db *gorm.DB, rdb *redis.Client) *MetricsService {
	ms := &MetricsService{
		logger: logger,
		db:     db,
		rdb:    rdb,
	}

	// 初始化Prometheus指标
	ms.initMetrics()

	return ms
}

func (ms *MetricsService) initMetrics() {
	// HTTP指标
	ms.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	ms.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求持续时间（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	ms.httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "当前正在处理的HTTP请求数",
		},
	)

	// 数据库指标
	ms.dbConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "数据库连接状态",
		},
		[]string{"database", "state"}, // state: open, idle, in_use
	)

	ms.dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "数据库查询持续时间（秒）",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"database", "operation"},
	)

	// Redis指标
	ms.redisConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_connections",
			Help: "Redis连接状态",
		},
		[]string{"state"}, // state: total, idle, active
	)

	ms.redisCommandDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_command_duration_seconds",
			Help:    "Redis命令执行持续时间（秒）",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"command"},
	)

	// 应用指标
	ms.activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "当前活跃连接数",
		},
	)

	ms.businessMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "business_metrics",
			Help: "业务指标",
		},
		[]string{"metric_name"},
	)
}

// HTTPMetricsMiddleware HTTP指标中间件
func (ms *MetricsService) HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 增加正在处理的请求计数
		ms.httpRequestsInFlight.Inc()
		defer ms.httpRequestsInFlight.Dec()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		ms.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		ms.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// DatabaseMetricsCallback 数据库指标回调
func (ms *MetricsService) DatabaseMetricsCallback() {
	// 定期更新数据库连接池指标
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ms.updateDatabaseMetrics()
		}
	}()
}

func (ms *MetricsService) updateDatabaseMetrics() {
	if ms.db != nil {
		sqlDB, err := ms.db.DB()
		if err != nil {
			ms.logger.Error("Failed to get SQL DB for metrics", zap.Error(err))
			return
		}

		stats := sqlDB.Stats()
		ms.dbConnections.WithLabelValues("postgres", "open").Set(float64(stats.OpenConnections))
		ms.dbConnections.WithLabelValues("postgres", "idle").Set(float64(stats.Idle))
		ms.dbConnections.WithLabelValues("postgres", "in_use").Set(float64(stats.InUse))
	}

	// 更新Redis连接指标
	if ms.rdb != nil {
		stats := ms.rdb.PoolStats()
		ms.redisConnections.WithLabelValues("total").Set(float64(stats.TotalConns))
		ms.redisConnections.WithLabelValues("idle").Set(float64(stats.IdleConns))
		ms.redisConnections.WithLabelValues("active").Set(float64(stats.TotalConns - stats.IdleConns))
	}
}

// RecordDatabaseQuery 记录数据库查询指标
func (ms *MetricsService) RecordDatabaseQuery(operation string, duration time.Duration) {
	ms.dbQueryDuration.WithLabelValues("postgres", operation).Observe(duration.Seconds())
}

// RecordRedisCommand 记录Redis命令指标
func (ms *MetricsService) RecordRedisCommand(command string, duration time.Duration) {
	ms.redisCommandDuration.WithLabelValues(command).Observe(duration.Seconds())
}

// UpdateBusinessMetrics 更新业务指标
func (ms *MetricsService) UpdateBusinessMetrics(ctx context.Context) {
	// 定期更新业务指标
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ms.collectBusinessMetrics(ctx)
		}
	}()
}

func (ms *MetricsService) collectBusinessMetrics(ctx context.Context) {
	// 获取用户总数
	var totalUsers int64
	if err := ms.db.WithContext(ctx).Model(&struct{}{}).
		Table("users").Count(&totalUsers).Error; err != nil {
		ms.logger.Error("Failed to count total users", zap.Error(err))
	} else {
		ms.businessMetrics.WithLabelValues("total_users").Set(float64(totalUsers))
	}

	// 获取24小时活跃用户数
	var activeUsers24h int64
	if err := ms.db.WithContext(ctx).Model(&struct{}{}).
		Table("users").
		Where("last_login_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&activeUsers24h).Error; err != nil {
		ms.logger.Error("Failed to count active users 24h", zap.Error(err))
	} else {
		ms.businessMetrics.WithLabelValues("active_users_24h").Set(float64(activeUsers24h))
	}

	// 获取24小时内创建的书签数
	var bookmarks24h int64
	if err := ms.db.WithContext(ctx).Model(&struct{}{}).
		Table("bookmarks").
		Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&bookmarks24h).Error; err != nil {
		ms.logger.Error("Failed to count bookmarks 24h", zap.Error(err))
	} else {
		ms.businessMetrics.WithLabelValues("bookmarks_created_24h").Set(float64(bookmarks24h))
	}

	// 获取24小时内的同步操作数
	var syncOps24h int64
	if err := ms.db.WithContext(ctx).Model(&struct{}{}).
		Table("sync_events").
		Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&syncOps24h).Error; err != nil {
		ms.logger.Error("Failed to count sync operations 24h", zap.Error(err))
	} else {
		ms.businessMetrics.WithLabelValues("sync_operations_24h").Set(float64(syncOps24h))
	}
}

// SetActiveConnections 设置活跃连接数
func (ms *MetricsService) SetActiveConnections(count float64) {
	ms.activeConnections.Set(count)
}

// MetricsHandler 返回Prometheus metrics处理器
func (ms *MetricsService) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// DatabaseQueryMiddleware 数据库查询中间件
type DatabaseQueryMiddleware struct {
	metricsService *MetricsService
}

// NewDatabaseQueryMiddleware 创建数据库查询中间件
func NewDatabaseQueryMiddleware(ms *MetricsService) *DatabaseQueryMiddleware {
	return &DatabaseQueryMiddleware{
		metricsService: ms,
	}
}

// WrapQuery 包装数据库查询以记录指标
func (m *DatabaseQueryMiddleware) WrapQuery(operation string, query func() error) error {
	start := time.Now()
	err := query()
	duration := time.Since(start)

	m.metricsService.RecordDatabaseQuery(operation, duration)
	return err
}

// RedisCommandMiddleware Redis命令中间件
type RedisCommandMiddleware struct {
	metricsService *MetricsService
}

// NewRedisCommandMiddleware 创建Redis命令中间件
func NewRedisCommandMiddleware(ms *MetricsService) *RedisCommandMiddleware {
	return &RedisCommandMiddleware{
		metricsService: ms,
	}
}

// WrapCommand 包装Redis命令以记录指标
func (m *RedisCommandMiddleware) WrapCommand(command string, cmd func() error) error {
	start := time.Now()
	err := cmd()
	duration := time.Since(start)

	m.metricsService.RecordRedisCommand(command, duration)
	return err
}