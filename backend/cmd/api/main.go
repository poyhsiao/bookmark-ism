package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/internal/server"
	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/logger"
	"bookmark-sync-service/backend/pkg/redis"
	"bookmark-sync-service/backend/pkg/search"
	"bookmark-sync-service/backend/pkg/storage"
	"bookmark-sync-service/backend/pkg/supabase"

	"go.uber.org/zap"
)

// convertStorageConfig converts config.StorageConfig to storage.Config
func convertStorageConfig(cfg config.StorageConfig) storage.Config {
	return storage.Config{
		Endpoint:        cfg.Endpoint,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		BucketName:      cfg.BucketName,
		UseSSL:          cfg.UseSSL,
	}
}

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

	// Initialize MinIO storage client
	storageClient, err := storage.NewClient(convertStorageConfig(cfg.Storage))
	if err != nil {
		logger.Fatal("Failed to connect to MinIO", zap.Error(err))
	}

	// Ensure storage bucket exists
	if err := storageClient.EnsureBucketExists(context.Background()); err != nil {
		logger.Fatal("Failed to ensure storage bucket exists", zap.Error(err))
	}

	// Initialize Typesense search client
	searchClient, err := search.NewClient(cfg.Search)
	if err != nil {
		logger.Fatal("Failed to connect to Typesense", zap.Error(err))
	}

	// Run database migrations
	if err := database.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize server
	srv := server.NewServer(cfg, db, redisClient, supabaseClient, storageClient, searchClient, logger)

	// Start server in a goroutine
	go func() {
		logger.Info("Starting API server", zap.String("port", cfg.Server.Port))
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
