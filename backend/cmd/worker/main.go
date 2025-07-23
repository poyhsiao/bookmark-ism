package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/logger"
	"bookmark-sync-service/backend/pkg/redis"
	"bookmark-sync-service/backend/pkg/supabase"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize Redis client
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient(cfg.Supabase)
	if err != nil {
		logger.Fatal("Failed to connect to Supabase", zap.Error(err))
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	go runLinkChecker(ctx, db, logger)
	go runCleanupJob(ctx, db, redisClient, logger)

	logger.Info("Worker service started")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker service...")
	cancel()
	logger.Info("Worker service exited")
}

// runLinkChecker periodically checks bookmarked links for validity
func runLinkChecker(ctx context.Context, db *gorm.DB, logger *zap.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	logger.Info("Starting link checker worker")

	for {
		select {
		case <-ticker.C:
			logger.Info("Running link check job")
			// TODO: Implement link checking logic
		case <-ctx.Done():
			logger.Info("Link checker worker stopped")
			return
		}
	}
}

// runCleanupJob periodically cleans up expired data
func runCleanupJob(ctx context.Context, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	logger.Info("Starting cleanup worker")

	for {
		select {
		case <-ticker.C:
			logger.Info("Running cleanup job")
			// TODO: Implement cleanup logic for expired tokens, temporary data, etc.
		case <-ctx.Done():
			logger.Info("Cleanup worker stopped")
			return
		}
	}
}
