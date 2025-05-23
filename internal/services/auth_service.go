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

// Генерация авторизационного кода (пока без client_id)
func (s *AuthService) GenerateAuthCode(userID, redirectURI, provider, providerID string) (string, error) {
	code := generateRandomCode()
	accessToken := "temporary" // заглушка до этапа получения токена PostToken()

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

	// Логируем db для отладки
	if s.db == nil {
		logrus.Infof("db!!!: %+v", s.db)
	}

	// Выполняем запрос с добавлением provider
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

// Получение и проверка валидности кода
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

	// Проверка на истечение срока действия
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("authorization code expired")
	}

	// Удаление использованного кода
	delQuery := `DELETE FROM oauth_sessions WHERE code = $1`
	_, _ = s.db.Exec(context.Background(), delQuery, code)

	return &authCode, nil
}

// Генерация случайного кода
func generateRandomCode() string {
	return uuid.NewString()
}
