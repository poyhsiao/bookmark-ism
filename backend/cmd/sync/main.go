package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/logger"
	"bookmark-sync-service/backend/pkg/redis"
	"bookmark-sync-service/backend/pkg/supabase"
	"bookmark-sync-service/backend/pkg/websocket"

	"go.uber.org/zap"
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

	// Create WebSocket hub
	wsHub := websocket.NewHub(redisClient, logger)

	// Start WebSocket hub
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wsHub.Run(ctx)

	// Subscribe to Redis channels for sync events
	pubsub := redisClient.Subscribe(ctx, "sync:events")
	defer pubsub.Close()

	// Start sync worker
	go func() {
		logger.Info("Starting sync worker")
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				logger.Error("Error receiving message", zap.Error(err))
				continue
			}

			logger.Info("Received sync event", zap.String("channel", msg.Channel), zap.String("payload", msg.Payload))
			// TODO: Process sync event
		}
	}()

	logger.Info("Sync service started")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down sync service...")
	cancel()
	logger.Info("Sync service exited")
}
