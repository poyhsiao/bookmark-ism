package main

import (
	"flag"
	"fmt"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Parse command line arguments
	var action string
	flag.StringVar(&action, "action", "up", "Migration action: up, down, reset, status")
	flag.Parse()

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

	// Run migrations based on action
	switch action {
	case "up":
		if err := runMigrationsUp(db, logger); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Migrations applied successfully")
	case "down":
		if err := runMigrationsDown(db, logger); err != nil {
			logger.Fatal("Failed to rollback migrations", zap.Error(err))
		}
		logger.Info("Migrations rolled back successfully")
	case "reset":
		if err := resetMigrations(db, logger); err != nil {
			logger.Fatal("Failed to reset migrations", zap.Error(err))
		}
		logger.Info("Migrations reset successfully")
	case "status":
		if err := showMigrationStatus(db, logger); err != nil {
			logger.Fatal("Failed to show migration status", zap.Error(err))
		}
	default:
		logger.Fatal("Invalid action. Use: up, down, reset, or status")
	}
}

// runMigrationsUp applies all pending migrations
func runMigrationsUp(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Running migrations up")

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// Run auto migrations for all models
	if err := database.AutoMigrate(db); err != nil {
		return err
	}

	return nil
}

// runMigrationsDown rolls back the last migration
func runMigrationsDown(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Running migrations down")
	// TODO: Implement proper migration rollback
	return fmt.Errorf("migration rollback not implemented yet")
}

// resetMigrations drops all tables and reapplies migrations
func resetMigrations(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Resetting migrations")

	// Get database connection
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Confirm reset
	fmt.Print("This will drop all tables and data. Are you sure? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		return fmt.Errorf("migration reset cancelled")
	}

	// Drop all tables
	if err := db.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE SCHEMA public").Error; err != nil {
		return err
	}

	// Run migrations
	return runMigrationsUp(db, logger)
}

// showMigrationStatus shows the current migration status
func showMigrationStatus(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Showing migration status")

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// Check if all models are migrated
	if err := db.Exec("SELECT 1 FROM users LIMIT 1").Error; err != nil {
		fmt.Println("Migration status: Not all tables exist")
		return nil
	}

	fmt.Println("Migration status: All tables exist")
	return nil
}

// createMigrationsTable creates the migrations table if it doesn't exist
func createMigrationsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}
