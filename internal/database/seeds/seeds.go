package seeds

import (
	"log"

	"gorm.io/gorm"
)

func SeedAll(db *gorm.DB) {
    log.Println("🌱 Starting user seeds...")
    SeedUsers(db) // run seeds
    log.Println("✅ User seeding completed successfully!")
}
