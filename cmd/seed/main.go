package main

import (
    "log"
	"github.com/vitalii-q/selena-users-service/internal/database/seeds"
	"github.com/vitalii-q/selena-users-service/internal/database"
	"github.com/vitalii-q/selena-users-service/internal/models"

	"github.com/joho/godotenv"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// run seeds: docker exec -it users-service_dev go run cmd/seed/main.go
func main() {
	// downloud .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found, relying on environment variables")
	}

    dsn := database.GetDatabaseURL()
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to DB: %v", err)
    }

    // Авто-маршрутизация моделей (чтобы таблицы были на месте)
    db.AutoMigrate(&models.User{})

    // run seeds
    seeds.SeedAll(db)
}
