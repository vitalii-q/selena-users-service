package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthCode struct {
    Code        string
    UserID      uuid.UUID
    ClientID    string
    RedirectURI string
    ExpiresAt   time.Time
}