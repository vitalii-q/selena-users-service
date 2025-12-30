package services

import (
	"context"
	"database/sql"
	"errors"

	//"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	//"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/pashagolub/pgxmock/v2"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

// UserServiceImpl - implementation of the user service
type UserServiceImpl struct {
	db db_interface
	passwordHasher utils.PasswordHasher
	hotelClient *external_services.HotelServiceClient
}

// NewUserServiceImpl - конструктор UserServiceImpl
func NewUserServiceImpl(
	db db_interface, 
	passwordHasher utils.PasswordHasher, 
	hotelClient *external_services.HotelServiceClient,
) *UserServiceImpl {
	return &UserServiceImpl{
		db: db, 
		passwordHasher: passwordHasher,
		hotelClient: hotelClient,
	}
}

// HotelClient — возвращает клиент для работы с отелями
func (s *UserServiceImpl) HotelClient() *external_services.HotelServiceClient {
	return s.hotelClient
}

// CreateUser - создание нового пользователя
func (s *UserServiceImpl) CreateUser(user models.User) (models.User, error) {
	// hash the password before saving it
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

// GetUser - getting a user by UUID
func (s *UserServiceImpl) GetUser(id uuid.UUID) (models.User, error) {
	var user models.User
	var gender, countryID, cityID sql.NullString

	query := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			role,
			birth,
			gender,
			country_id,
			city_id,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Role,
		&user.Birth,
		&gender,
		&countryID,
		&cityID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	// Validation
	if gender.Valid {
		user.Gender = &gender.String
	}
	if countryID.Valid {
		id, _ := uuid.Parse(countryID.String)
		user.CountryID = &id
	}
	if cityID.Valid {
		id, _ := uuid.Parse(cityID.String)
		user.CityID = &id
	}

	return user, nil
}

// GetUserByEmail - receiving a user by email
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

// UpdateUser - updating user data
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

// DeleteUser - deleting a user by ID
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

// GetAllUsers — returns all users
func (s *UserServiceImpl) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			role,
			birth,
			gender,
			country_id,
			city_id,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	ctx := context.Background()

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var user models.User
		var gender, countryID, cityID sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Role,
			&user.Birth,
			&gender,
			&countryID,
			&cityID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		// Validation
		if gender.Valid {
			user.Gender = &gender.String
		}
		if countryID.Valid {
			id, _ := uuid.Parse(countryID.String)
			user.CountryID = &id
		}
		if cityID.Valid {
			id, _ := uuid.Parse(cityID.String)
			user.CityID = &id
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

