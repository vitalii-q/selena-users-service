package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/vitalii-q/selena-users-service/internal/models"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`

	Birth *string `json:"birth"` // omitempty - not show if null

	Gender *string `json:"gender"`

	CountryID *uuid.UUID `json:"country_id"`
	Country   *string    `json:"country"`

	CityID *uuid.UUID `json:"city_id"`
	City   *string    `json:"city"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func MapUser(user models.User) UserResponse {
	response := UserResponse{
		ID: user.ID, FirstName: user.FirstName, LastName: user.LastName,
		Email: user.Email, Role: user.Role, Gender: user.Gender,
		CountryID: user.CountryID, CityID: user.CityID,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
	if user.Birth != nil {
		birth := user.Birth.Format("2006-01-02")
		response.Birth = &birth
	}
	return response
}
