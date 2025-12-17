package seeds

import (
	"log"
	"time"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/utils"
	"gorm.io/gorm"
)

// SeedUsers fills the users table with test data
func SeedUsers(db *gorm.DB) {
	// Create a hasher
	hasher := utils.NewBcryptHasher()

	users := []models.User{
		{
			Email:     "admin@mail.com",
			Password:  "password",
			FirstName: "admin",
			LastName:  "admin",
			Role:      "admin",
			Birth:     nil, // можно оставить пустым
			Gender:    "male",
			Country:   "Germany",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "user@mail.com",
			Password:  "password",
			FirstName: "user",
			LastName:  "user",
			Role:      "user",
			Birth:     ptrTime(time.Date(1990, time.May, 21, 0, 0, 0, 0, time.UTC)),
			Gender:    "male",
			Country:   "Germany",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "user2@mail.com",
			Password:  "password",
			FirstName: "user2",
			LastName:  "user2",
			Role:      "user",
			Birth:     ptrTime(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
			Gender:    "male",
			Country:   "France",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "user3@mail.com",
			Password:  "password",
			FirstName: "user3",
			LastName:  "user3",
			Role:      "user",
			Birth:     ptrTime(time.Date(1985, time.December, 31, 0, 0, 0, 0, time.UTC)),
			Gender:    "female",
			Country:   "France",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, user := range users {
		var existing models.User

		// Check if a user with this email address exists
		err := db.Where("email = ?", user.Email).First(&existing).Error

		if err == nil {
			log.Printf("User %s already exists, skipping...", user.Email)
			continue
		}

		if err != gorm.ErrRecordNotFound {
			log.Printf("Failed to check user %s: %v", user.Email, err)
			continue
		}

		// Hash the password before inserting it
		hashedPassword, err := hasher.HashPassword(user.Password)
		if err != nil {
			log.Printf("Failed to hash password for user %s: %v", user.Email, err)
			continue
		}

		err = db.Model(&models.User{}).Create(map[string]interface{}{
			"email":         user.Email,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"password_hash": hashedPassword,
			"role":          user.Role,
			"birth":   		 user.Birth,
			"gender":  		 user.Gender,
			"country": 		 user.Country,
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
		}).Error

		if err != nil {
			log.Printf("Failed to seed user %s: %v", user.Email, err)
		} else {
			log.Printf("User %s seeded successfully", user.Email)
		}
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
