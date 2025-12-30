package dto

import "github.com/google/uuid"

type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`

	Birth 	  *string    `json:"birth"`   // omitempty - not show if null

	Gender    *string    `json:"gender"`

	CountryID *uuid.UUID `json:"country_id"`
	Country   *string     `json:"country"`

	CityID    *uuid.UUID `json:"city_id"`
	City      *string     `json:"city"`

	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}
