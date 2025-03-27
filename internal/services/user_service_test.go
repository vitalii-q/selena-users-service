package services

import (
	//"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
