package services

import (
	"context"
	"errors"

	//"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	//"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/pashagolub/pgxmock/v2"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

// UserServiceImpl - реализация сервиса пользователей
type UserServiceImpl struct {
	db db_interface
	passwordHasher utils.PasswordHasher
}

// NewUserServiceImpl - конструктор UserServiceImpl
func NewUserServiceImpl(db db_interface, passwordHasher utils.PasswordHasher) *UserServiceImpl {
	return &UserServiceImpl{
		db: db, 
		passwordHasher: passwordHasher,
	}
}

// CreateUser - создание нового пользователя
func (s *UserServiceImpl) CreateUser(user models.User) (models.User, error) {
	// Хешируем пароль перед сохранением
	hashedPassword, err := s.passwordHasher.HashPassword(user.Password)
	if err != nil {
		return models.User{}, err
	}

	query := `INSERT INTO users (first_name, last_name, email, password_hash, role, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, created_at, updated_at`

	err = s.db.QueryRow(context.Background(), query,
		user.FirstName, user.LastName, user.Email, hashedPassword, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// GetUser - получение пользователя по UUID
func (s *UserServiceImpl) GetUser(id uuid.UUID) (models.User, error) {
	var user models.User
	query := `SELECT id, first_name, last_name, email, role, created_at, updated_at, deleted_at
		  FROM users WHERE id = $1 AND deleted_at IS NULL`

	err := s.db.QueryRow(context.Background(), query, id.String()).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email,
		&user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

// GetUserByEmail - получение пользователя по email
func (s *UserServiceImpl) GetUserByEmail(email string) (models.UserAuth, error) {
	var user models.UserAuth
	query := `SELECT id, email, password_hash FROM users WHERE email = $1 AND deleted_at IS NULL`

	err := s.db.QueryRow(context.Background(), query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserAuth{}, errors.New("user not found")
		}
		return models.UserAuth{}, err
	}

	return user, nil
}

// UpdateUser - обновление данных пользователя
func (s *UserServiceImpl) UpdateUser(id uuid.UUID, updatedUser models.User) (models.User, error) {
	query := `UPDATE users 
			  SET first_name = $1, last_name = $2, email = $3, updated_at = NOW()
			  WHERE id = $4 RETURNING updated_at`

	err := s.db.QueryRow(context.Background(), query,
		updatedUser.FirstName, updatedUser.LastName, updatedUser.Email, id).
		Scan(&updatedUser.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	updatedUser.ID = id
	return updatedUser, nil
}

// DeleteUser - удаление пользователя по ID
func (s *UserServiceImpl) DeleteUser(id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
	result, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found") 
	}

	return nil
}
