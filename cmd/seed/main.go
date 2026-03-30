package main

import (
	"log"

	"github.com/vitalii-q/selena-users-service/internal/config"
	"github.com/vitalii-q/selena-users-service/internal/database"
	"github.com/vitalii-q/selena-users-service/internal/database/seeds"
	"github.com/vitalii-q/selena-users-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Run seeds: docker exec -it users-service go run cmd/seed/main.go
// Run seeds in cloud: docker exec -it users-service /app/bin/seed
//
// The order of seeding: hotels, locations (hotels-service) -> users (users-service) -> bookings (bookings-service)
func main() {
    dsn := database.GetDatabaseURL(config.LoadEnv())
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to DB: %v", err)
    }

    // Auto-routing of models (so that tables are in place)
    db.AutoMigrate(&models.User{})

    // run seeds
    seeds.SeedAll(db)
}
