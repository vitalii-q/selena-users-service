package handlers_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

var dbPool *pgxpool.Pool

func TestCreateUser(t *testing.T) {
	//logrus.Infof("dbPool3: %#v", dbPool)

	// Создаем объект passwordHasher (можно использовать реальную реализацию)
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)

	// Создаем нового пользователя
	user := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		Role:      "user",
	}

	// Выполняем создание пользователя через сервис
	createdUser, err := userService.CreateUser(user)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)

	// Проверяем, что пользователь был создан
	assert.NotNil(t, createdUser.ID)
	assert.Equal(t, user.FirstName, createdUser.FirstName)
	assert.Equal(t, user.LastName, createdUser.LastName)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Equal(t, user.Role, createdUser.Role)
}