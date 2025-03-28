package services

import (
	//"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	//"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vitalii-q/selena-users-service/internal/models"
)

// Тест для CreateUser
func TestCreateUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close() // закрывает ресурсы, связанные с мок-объектом после выполнения func TestCreateUser

	// Данные пользователя
	newUser := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		Password:  "hashedpassword",
		Role:      "user",
	}

	createdAt := time.Now()
	updatedAt := createdAt
	userID := uuid.New()

	// Ожидаем, что будет вызван SQL-запрос с такими параметрами
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(newUser.FirstName, newUser.LastName, newUser.Email, newUser.Password, newUser.Role).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(userID, createdAt, updatedAt))

	// Создаем сервис с мокнутым соединением
	userService := NewUserService(mock)

	// Запускаем тестируемый метод
	createdUser, err := userService.CreateUser(newUser)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)
	assert.Equal(t, userID, createdUser.ID)
	assert.Equal(t, createdAt, createdUser.CreatedAt)
	assert.Equal(t, updatedAt, createdUser.UpdatedAt)

	// Проверяем, что все ожидания моков были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	// Данные пользователя
	userID := uuid.New()
	expectedUser := models.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Ожидаем, что будет вызван SQL-запрос с такими параметрами
	mock.ExpectQuery(`SELECT id, first_name, last_name, email, role, created_at, updated_at, deleted_at FROM users WHERE id = \$1`).
		WithArgs(userID.String()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "first_name", "last_name", "email", "role", "created_at", "updated_at", "deleted_at"}).
			AddRow(expectedUser.ID, expectedUser.FirstName, expectedUser.LastName, expectedUser.Email, expectedUser.Role, expectedUser.CreatedAt, expectedUser.UpdatedAt, nil))

	// Создаем сервис с мокнутым соединением
	userService := NewUserService(mock)

	// Запускаем тестируемый метод
	user, err := userService.GetUser(userID)

	// Проверяем
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	// Проверяем, что все ожидания моков были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUser_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()

	// Ожидаем, что запрос не найдет пользователя
	mock.ExpectQuery(`SELECT id, first_name, last_name, email, role, created_at, updated_at, deleted_at FROM users WHERE id = \$1`).
		WithArgs(userID.String()).
		WillReturnError(pgx.ErrNoRows)

	userService := NewUserService(mock)

	_, err = userService.GetUser(userID)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	updatedUser := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "new_email@example.com",
	}
	updatedAt := time.Now()

	mock.ExpectQuery(`UPDATE users SET first_name = \$1, last_name = \$2, email = \$3, updated_at = NOW\(\) WHERE id = \$4 RETURNING updated_at`).
		WithArgs(updatedUser.FirstName, updatedUser.LastName, updatedUser.Email, userID).
		WillReturnRows(pgxmock.NewRows([]string{"updated_at"}).AddRow(updatedAt))

	userService := NewUserService(mock)

	result, err := userService.UpdateUser(userID, updatedUser)

	assert.NoError(t, err)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, updatedAt, result.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUser_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	updatedUser := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "new_email@example.com",
	}

	mock.ExpectQuery(`UPDATE users SET first_name = \$1, last_name = \$2, email = \$3, updated_at = NOW\(\) WHERE id = \$4 RETURNING updated_at`).
		WithArgs(updatedUser.FirstName, updatedUser.LastName, updatedUser.Email, userID).
		WillReturnError(pgx.ErrNoRows)

	userService := NewUserService(mock)

	result, err := userService.UpdateUser(userID, updatedUser)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Equal(t, models.User{}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()

	mock.ExpectExec(`UPDATE users SET deleted_at = NOW\(\) WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	userService := NewUserService(mock)

	err = userService.DeleteUser(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()

	mock.ExpectExec(`UPDATE users SET deleted_at = NOW\(\) WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	userService := NewUserService(mock)

	err = userService.DeleteUser(userID)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
