package main

import (
	"log"

	"github.com/vitalii-q/selena-users-service/internal/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Clean users table: docker exec -it users-service go run cmd/clean/main.go
func main() {
	dsn := database.GetDatabaseURL()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	log.Println("🧹 Cleaning users table...")

	err = db.Exec(`
		TRUNCATE TABLE users
		RESTART IDENTITY
		CASCADE
	`).Error

	if err != nil {
		log.Fatalf("Failed to clean users table: %v", err)
	}

	log.Println("✅ Users table cleaned successfully")
}