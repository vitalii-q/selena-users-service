package services

import (
	"context"
	"time"
	"github.com/google/uuid"

	"github.com/vitalii-q/selena-users-service/internal/models"
)

type AuthService struct {
	db db_interface
}

func NewAuthService(db db_interface) *AuthService {
	return &AuthService{
		db: db,
	}
}

func (s *AuthService) GenerateAuthCode(userID, redirectURI string) (string, error) {
	code := generateRandomCode() // можешь использовать uuid.NewString() или свою генерацию

	query := `INSERT INTO oauth_sessions (code, user_id, redirect_uri, expires_at)
			  VALUES ($1, $2, $3, $4)`

	_, err := s.db.Exec(context.Background(), query, code, userID, redirectURI, time.Now().Add(5*time.Minute))
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *AuthService) GetAuthCode(code string) (*models.AuthCode, error) {
	var authCode models.AuthCode

	query := `SELECT code, user_id, client_id, redirect_uri, expires_at FROM oauth_sessions WHERE code = $1`

	err := s.db.QueryRow(context.Background(), query, code).Scan(
		&authCode.Code, &authCode.UserID, &authCode.ClientID, &authCode.RedirectURI, &authCode.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	return &authCode, nil
}

func generateRandomCode() string {
	return uuid.NewString()
}