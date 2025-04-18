package models

import "time"

type AuthCode struct {
    Code        string
    UserID      string
    ClientID    string
    RedirectURI string
    ExpiresAt   time.Time
}