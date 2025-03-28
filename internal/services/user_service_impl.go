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
)

// UserServiceImpl - реализация сервиса пользователей
type UserServiceImpl struct {
	db db_interface
}

// NewUserServiceImpl - конструктор UserServiceImpl
func NewUserServiceImpl(db db_interface) *UserServiceImpl {
	return &UserServiceImpl{db: db}
}

// CreateUser - создание нового пользователя
func (s *UserServiceImpl) CreateUser(user models.User) (models.User, error) {
	query := `INSERT INTO users (first_name, last_name, email, password_hash, role, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(context.Background(), query,
		user.FirstName, user.LastName, user.Email, user.Password, user.Role).
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
			  FROM users WHERE id = $1`

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
	_, err := s.db.Exec(context.Background(), query, id)

	if err != nil {
		return err
	}

	return nil
}
