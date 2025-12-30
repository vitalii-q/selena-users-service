package models

import (
	"time"

	"github.com/google/uuid"
)

// User - user model
type User struct {
	ID           uuid.UUID  `json:"id,omitempty"`
	FirstName    string     `json:"first_name" validate:"required,min=2"`
	LastName     string     `json:"last_name" validate:"required,min=2"`
	Email        string     `json:"email" validate:"required,email"`
	Password     string     `json:"password,omitempty" validate:"required,min=6"`
	Role         string     `json:"role" validate:"required,oneof=admin user"`
	Birth        *time.Time `json:"birth,omitempty"`       // nullable
	Gender  	 *string 	`json:"gender"`                // nullable
	CountryID 	 *uuid.UUID `json:"country_id"`            // nullable
	CityID    	 *uuid.UUID `json:"city_id"`               // nullable
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`  // nullable
}
