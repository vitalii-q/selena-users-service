// internal/models/auth.go
package models

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	Role      string `json:"role" validate:"required,oneof=admin user"`
}

type UserAuth struct {
	ID           uuid.UUID `json:"id,omitempty"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
}
