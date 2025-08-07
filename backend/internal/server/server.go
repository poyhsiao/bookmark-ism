package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"bookmark-sync-service/backend/internal/auth"
	"bookmark-sync-service/backend/internal/bookmark"
	"bookmark-sync-service/backend/internal/collection"
	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/internal/content"
	import_export "bookmark-sync-service/backend/internal/import"
	"bookmark-sync-service/backend/internal/monitoring"
	"bookmark-sync-service/backend/internal/search"
	"bookmark-sync-service/backend/internal/user"
	"bookmark-sync-service/backend/pkg/middleware"
	"bookmark-sync-service/backend/pkg/redis"
	searchpkg "bookmark-sync-service/backend/pkg/search"
	"bookmark-sync-service/backend/pkg/storage"
	"bookmark-sync-service/backend/pkg/supabase"
	"bookmark-sync-service/backend/pkg/utils"
	"bookmark-sync-service/backend/pkg/websocket"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server represents the HTTP server
type Server struct {
	config              *config.Config
	db                  *gorm.DB
	redisClient         *redis.Client
	supabaseClient      *supabase.Client
	storageClient       *storage.Client
	searchClient        *searchpkg.Client
	logger              *zap.Logger
	router              *gin.Engine
	httpServer          *http.Server
	wsHub               *websocket.Hub
	authHandler         *auth.Handler
	userHandler         *user.Handler
	bookmarkHandler     *bookmark.Handlers
	collectionHandler   *collection.Handler
	searchHandler       *search.Handlers
	importExportHandler *import_export.Handlers
	contentHandler      *content.Handler
	monitoringHandler   *monitoring.Handler
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, supabaseClient *supabase.Client, storageClient *storage.Client, searchClient *searchpkg.Client, logger *zap.Logger) *Server {
	// Set Gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create WebSocket hub
	wsHub := websocket.NewHub(redisClient, logger)

	// Create auth service and handler
	authService := auth.NewService(db, redisClient, supabaseClient, &cfg.JWT, logger)
	authHandler := auth.NewHandler(authService, logger)

	// Create user service and handler
	userService := user.NewService(db, storageClient, logger)
	userHandler := user.NewHandler(userService, logger)

	// Create bookmark service and handler
	bookmarkService := bookmark.NewService(db)
	bookmarkHandler := bookmark.NewHandlers(bookmarkService)

	// Create collection service and handler
	collectionService := collection.NewService(db)
	collectionHandler := collection.NewHandler(collectionService)

	// Create search service and handler
	searchService, err := search.NewService(cfg.Search)
	if err != nil {
		logger.Error("Failed to create search service", zap.Error(err))
		// Continue without search service for now
		searchService = nil
	}
	var searchHandler *search.Handlers
	if searchService != nil {
		searchHandler = search.NewHandlers(searchService)
	}

	// Create import/export service and handler
	importExportService := import_export.NewService(db)
	importExportHandler := import_export.NewHandlers(importExportService)

	// Create content service and handler
	contentService := content.NewService()
	contentHandler := content.NewHandler(contentService, cfg)

	// Create monitoring service and handler
	monitoringService := monitoring.NewService(db)
	monitoringHandler := monitoring.NewHandler(monitoringService)

	server := &Server{
		config:              cfg,
		db:                  db,
		redisClient:         redisClient,
		supabaseClient:      supabaseClient,
		storageClient:       storageClient,
		searchClient:        searchClient,
		logger:              logger,
		router:              gin.New(),
		wsHub:               wsHub,
		authHandler:         authHandler,
		userHandler:         userHandler,
		bookmarkHandler:     bookmarkHandler,
		collectionHandler:   collectionHandler,
		searchHandler:       searchHandler,
		importExportHandler: importExportHandler,
		contentHandler:      contentHandler,
		monitoringHandler:   monitoringHandler,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures middleware for the server
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Tracing middleware (includes request ID and structured logging)
	s.router.Use(utils.TracingMiddleware(s.logger))

	// CORS middleware
	s.router.Use(s.corsMiddleware())

	// Rate limiting middleware (placeholder for now)
	s.router.Use(s.rateLimitMiddleware())
}

// setupRoutes configures routes for the server
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", s.authHandler.Register)
			authGroup.POST("/login", s.authHandler.Login)
			authGroup.POST("/refresh", s.authHandler.RefreshToken)
			authGroup.POST("/reset", s.authHandler.ResetPassword)
			authGroup.POST("/validate", s.authHandler.ValidateToken)
		}

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(&s.config.JWT))
		{
			// Auth routes that require authentication
			protected.POST("/auth/logout", s.authHandler.Logout)
			protected.GET("/auth/profile", s.authHandler.GetProfile)

			// Register bookmark routes
			s.bookmarkHandler.RegisterRoutes(protected)

			// Register collection routes
			s.collectionHandler.RegisterRoutes(protected)

			// Register import/export routes
			s.importExportHandler.RegisterRoutes(protected)

			// Register content analysis routes
			s.contentHandler.RegisterRoutes(protected)

			// Register monitoring routes
			s.monitoringHandler.RegisterRoutes(protected)

			// Sync routes
			sync := protected.Group("/sync")
			{
				sync.GET("/changes", s.placeholder)
				sync.POST("/push", s.placeholder)
				sync.GET("/status", s.placeholder)
				sync.POST("/devices", s.placeholder)
				sync.GET("/devices", s.placeholder)
			}

			// User profile routes
			userGroup := protected.Group("/user")
			{
				userGroup.GET("/profile", s.userHandler.GetProfile)
				userGroup.PUT("/profile", s.userHandler.UpdateProfile)
				userGroup.GET("/preferences", s.userHandler.GetPreferences)
				userGroup.PUT("/preferences", s.userHandler.UpdatePreferences)
				userGroup.POST("/avatar", s.userHandler.UploadAvatar)
				userGroup.GET("/stats", s.userHandler.GetStats)
				userGroup.POST("/export", s.userHandler.ExportData)
				userGroup.DELETE("/account", s.userHandler.DeleteAccount)
			}

			// Storage routes
			storage := protected.Group("/storage")
			{
				storage.POST("/screenshots", s.placeholder)
				storage.GET("/screenshots/:id", s.placeholder)
				storage.POST("/files", s.placeholder)
				storage.GET("/files/:id", s.placeholder)
				storage.DELETE("/files/:id", s.placeholder)
			}
		}

		// Public routes (optional authentication)
		public := v1.Group("/")
		public.Use(middleware.OptionalAuthMiddleware(&s.config.JWT))
		{
			// Community routes
			community := public.Group("/community")
			{
				community.GET("/trending", s.placeholder)
				community.GET("/popular", s.placeholder)
				community.GET("/collections/:id", s.placeholder)
				community.GET("/users/:id", s.placeholder)
				community.POST("/follow/:id", s.placeholder)
				community.DELETE("/follow/:id", s.placeholder)
				community.GET("/feed", s.placeholder)
			}

			// Search routes
			if s.searchHandler != nil {
				s.searchHandler.RegisterRoutes(public)
			} else {
				search := public.Group("/search")
				{
					search.GET("/bookmarks", s.placeholder)
					search.GET("/collections", s.placeholder)
					search.GET("/users", s.placeholder)
					search.GET("/suggest", s.placeholder)
					search.GET("/tags", s.placeholder)
				}
			}
		}

		// WebSocket endpoint (requires authentication via query params)
		v1.GET("/sync/ws", s.wsHub.HandleWebSocket)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port),
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
	}

	// Start WebSocket hub in a separate goroutine
	go s.wsHub.Run(context.Background())

	s.logger.Info("Server starting",
		zap.String("address", s.httpServer.Addr),
		zap.String("environment", s.config.Server.Environment),
	)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Server shutting down...")
	return s.httpServer.Shutdown(ctx)
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	healthData := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"services":  gin.H{},
	}

	services := healthData["services"].(gin.H)

	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		services["database"] = gin.H{"status": "unhealthy", "error": "connection error"}
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Database connection error", nil)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		services["database"] = gin.H{"status": "unhealthy", "error": "ping failed"}
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Database is currently unavailable", nil)
		return
	}
	services["database"] = gin.H{"status": "healthy"}

	// Check Redis connection
	if err := s.redisClient.Ping(c.Request.Context()); err != nil {
		services["redis"] = gin.H{"status": "unhealthy", "error": "connection error"}
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Redis is currently unavailable", nil)
		return
	}
	services["redis"] = gin.H{"status": "healthy"}

	// Check Supabase connection
	if err := s.supabaseClient.HealthCheck(c.Request.Context()); err != nil {
		services["supabase"] = gin.H{"status": "unhealthy", "error": "connection error"}
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Supabase is currently unavailable", nil)
		return
	}
	services["supabase"] = gin.H{"status": "healthy"}

	// Check MinIO connection
	if err := s.storageClient.HealthCheck(c.Request.Context()); err != nil {
		services["storage"] = gin.H{"status": "unhealthy", "error": "connection error"}
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "MinIO is currently unavailable", nil)
		return
	}
	services["storage"] = gin.H{"status": "healthy"}

	// Check Typesense connection
	if err := s.searchClient.HealthCheck(c.Request.Context()); err != nil {
		services["search"] = gin.H{"status": "unhealthy", "error": "connection error"}
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Typesense is currently unavailable", nil)
		return
	}
	services["search"] = gin.H{"status": "healthy"}

	utils.SuccessResponse(c, healthData, "System is healthy")
}

// placeholder handler for routes not yet implemented
func (s *Server) placeholder(c *gin.Context) {
	utils.ErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		fmt.Sprintf("Endpoint %s %s is not yet implemented", c.Request.Method, c.Request.URL.Path), nil)
}

// rateLimitMiddleware provides basic rate limiting (placeholder implementation)
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting using Redis
		// For now, this is a placeholder that allows all requests
		c.Next()
	}
}

// corsMiddleware handles CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
