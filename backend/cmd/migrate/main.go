package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/database"
)

func main() {
	var direction = flag.String("direction", "up", "Migration direction: up or down")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	switch *direction {
	case "up":
		fmt.Println("Running database migrations...")
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("✅ Migrations completed successfully!")

	case "down":
		fmt.Println("⚠️  Rolling back migrations...")
		if err := database.Rollback(db); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		fmt.Println("✅ Rollback completed successfully!")

	case "seed":
		fmt.Println("Seeding database with test data...")
		if err := database.SeedTestData(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		fmt.Println("✅ Database seeded successfully!")

	default:
		fmt.Printf("Unknown direction: %s. Use 'up', 'down', or 'seed'\n", *direction)
		os.Exit(1)
	}
}
