package seeds

import (
	"log"

	"gorm.io/gorm"
)

// run seeds: docker exec -it users-service go run cmd/seed/main.go
//
// The order of seeding: hotels, locations (hotels-service) -> users (users-service) -> bookings (bookings-service)
func SeedAll(db *gorm.DB) {
    log.Println("🌱 Starting user seeds...")
    SeedUsers(db) // run seeds
    log.Println("✅ User seeding completed successfully!")
}
