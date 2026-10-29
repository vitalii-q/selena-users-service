package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vitalii-q/selena-users-service/internal/models"
)

type AuthService struct {
	db db_interface
}

func NewAuthService(db db_interface) *AuthService {
	return &AuthService{db: db}
}

// Generation of an authorization code (without client_id for now)
func (s *AuthService) GenerateAuthCode(userID, redirectURI, provider, providerID string) (string, error) {
	code := generateRandomCode()
	accessToken := "temporary" // stub before the token receipt stage PostToken()

	// Добавляем provider в запрос
	query := `
	INSERT INTO oauth_sessions (code, user_id, redirect_uri, provider, provider_id, access_token, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (provider, provider_id) DO UPDATE
	SET code = EXCLUDED.code,
	    redirect_uri = EXCLUDED.redirect_uri,
	    access_token = EXCLUDED.access_token,
	    expires_at = EXCLUDED.expires_at,
	    updated_at = NOW()
	`

	// Logging the db for debugging
	if s.db == nil {
		logrus.Infof("db!!!: %+v", s.db)
	}

	// Logging the db for debugging
	_, err := s.db.Exec(
		context.Background(), 
		query, 
		code, 
		userID, 
		redirectURI, 
		provider, 
		providerID, 
		accessToken,
		time.Now().Add(5*time.Minute),
	)
	if err != nil {
		return "", err
	}

	return code, nil
}

// Getting and verifying the validity of the code
func (s *AuthService) GetAuthCode(code string) (*models.AuthCode, error) {
	var authCode models.AuthCode

	query := `SELECT code, user_id, client_id, redirect_uri, expires_at
			  FROM oauth_sessions WHERE code = $1`

	err := s.db.QueryRow(context.Background(), query, code).Scan(
		&authCode.Code, &authCode.UserID, &authCode.ClientID,
		&authCode.RedirectURI, &authCode.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	// Checking for expiration
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("authorization code expired")
	}

	// Deleting the used code
	delQuery := `DELETE FROM oauth_sessions WHERE code = $1`
	_, _ = s.db.Exec(context.Background(), delQuery, code)

	return &authCode, nil
}

// Generating random code
func generateRandomCode() string {
	return uuid.NewString()
}
