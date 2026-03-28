package main

import (
	"log"
	//"os"

	"github.com/vitalii-q/selena-users-service/internal/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Clean users table: docker exec -it users-service go run cmd/clean/main.go
func main() {
	/*if os.Getenv("APP_ENV") == "production" {
		log.Fatal("Cleaning users table is not allowed in production")
	}*/

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