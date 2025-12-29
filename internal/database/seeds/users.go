package seeds

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/utils"
	"gorm.io/gorm"
)

// ----------------------------------------------------------------
// run seeds: docker exec -it users-service go run cmd/seed/main.go
// ----------------------------------------------------------------

type AgeRange struct {
	Min int
	Max int
}

// SeedUsers fills the users table with test data
func SeedUsers(db *gorm.DB) {
	hasher := utils.NewBcryptHasher()

	// -------------------------------
	// ❗ CLEARING THE TABLE
	// -------------------------------
	if err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error; err != nil {
	log.Fatalf("Failed to truncate users table: %v", err)}

	// CONFIGURATION
	totalUsers := 50

	ageRanges := []AgeRange{
		{Min: 18, Max: 25},
		{Min: 26, Max: 35},
		{Min: 36, Max: 50},
	}

	countries := []string{"Germany", "France", "Ukraine"} // TODO: get contries and cities from hotels-service
	genders := []string{"male", "female"}

	now := time.Now()

	var users []models.User
	var passwordHashes []string

	// ADMIN (static, included in batch)
	adminPasswordHash, err := hasher.HashPassword("password")
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}

	users = append(users, models.User{
		Email:     "admin@mail.com",
		Password:  "password",
		FirstName: "admin",
		LastName:  "admin",
		Role:      "admin",
		Birth:     nil,
		Gender:    "male",
		Country:   "Germany",
		CreatedAt: now,
		UpdatedAt: now,
	})

	passwordHashes = append(passwordHashes, adminPasswordHash)

	// GENERATED USERS
	usersPerRange := totalUsers / len(ageRanges)
	userIndex := 1

	for _, ageRange := range ageRanges {
		for i := 0; i < usersPerRange; i++ {
			age := randomInt(ageRange.Min, ageRange.Max)
			birth := birthDateFromAge(age)

			passwordHash, err := hasher.HashPassword("password")
			if err != nil {
				log.Printf("Failed to hash password: %v", err)
				continue
			}

			users = append(users, models.User{
				Email:     fmt.Sprintf("user%d@mail.com", userIndex),
				Password:  "password",
				FirstName: fmt.Sprintf("User%d", userIndex),
				LastName:  fmt.Sprintf("LastName%d", userIndex),
				Role:      "user",
				Birth:     birth,
				Gender:    genders[rand.Intn(len(genders))],
				Country:   countries[rand.Intn(len(countries))],
				CreatedAt: now,
				UpdatedAt: now,
			})

			passwordHashes = append(passwordHashes, passwordHash)
			userIndex++
		}
	}

	// ONE BATCH INSERT
	if len(users) > 0 {
		var insertData []map[string]interface{}
		for i, u := range users {
			insertData = append(insertData, map[string]interface{}{
				"email":         u.Email,
				"first_name":    u.FirstName,
				"last_name":     u.LastName,
				"password_hash": passwordHashes[i],
				"role":          u.Role,
				"birth":         u.Birth,
				"gender":        u.Gender,
				"country":       u.Country,
				"created_at":    u.CreatedAt,
				"updated_at":    u.UpdatedAt,
			})
		}

		err := db.Model(&models.User{}).Create(&insertData).Error
		if err != nil {
			log.Printf("Failed to batch insert users: %v", err)
		} else {
			log.Printf("%d users seeded successfully (single batch insert)", len(users))
		}
	}
}

func randomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func birthDateFromAge(age int) *time.Time {
	now := time.Now()
	year := now.Year() - age
	month := time.Month(rand.Intn(12) + 1)
	day := rand.Intn(28) + 1

	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t
}
