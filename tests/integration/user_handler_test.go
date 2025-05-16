package integration_tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	//"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

// Изолированный роутинг что бы не затрагивать OAuth, /protected, /test
func setupTestRouter(userHandler *handlers.UserHandler) *gin.Engine {
	router := gin.Default()
	router.GET("/users/:id", userHandler.GetUserHandler)
	router.PUT("/users/:id", userHandler.UpdateUserHandler)
	router.DELETE("/users/:id", userHandler.DeleteUserHandler)
	return router
}

func TestCreateUser(t *testing.T) {
	//logrus.Infof("Test dbPool: %#v", dbPool)

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

func TestCreateUserWithEmptyFields(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	payload := `{"first_name": "", "last_name": "", "email": "", "password": "", "role": ""}`
	req, _ := http.NewRequest("POST", "/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.POST("/users", userHandler.CreateUserHandler)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Validation failed")
}

func TestCreateUserWithDuplicateEmail(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	user := models.User{
		FirstName: "Alex", LastName: "Jones",
		Email: "alex.jones@example.com", Password: "pass123", Role: "user",
	}
	_, _ = userService.CreateUser(user)

	payload := `{"first_name": "Another", "last_name": "User", "email": "alex.jones@example.com", "password": "anotherpass", "role": "user"}`
	req, _ := http.NewRequest("POST", "/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.POST("/users", userHandler.CreateUserHandler)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "duplicate")
}

func TestGetUserHandler(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)

	// Создаем тестового пользователя
	user := models.User{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Password:  "securepassword",
		Role:      "user",
	}
	createdUser, err := userService.CreateUser(user)
	require.NoError(t, err)

	router := setupTestRouter(userHandler)

	req, _ := http.NewRequest("GET", "/users/"+createdUser.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var returnedUser models.User
	err = json.Unmarshal(w.Body.Bytes(), &returnedUser)
	require.NoError(t, err)

	assert.Equal(t, createdUser.ID, returnedUser.ID)
	assert.Equal(t, user.Email, returnedUser.Email)
}

func TestGetNonExistingUser(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	nonExistingID := uuid.New()

	req, _ := http.NewRequest("GET", "/users/"+nonExistingID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserWithInvalidUUID(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	req, _ := http.NewRequest("GET", "/users/not-a-valid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid UUID format")
}

func TestUpdateUserHandler(t *testing.T) {
	// Создаем сервис и handler
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)

	// Создаем пользователя
	user := models.User{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice.smith@example.com",
		Password:  "password123",
		Role:      "user",
	}
	createdUser, _ := userService.CreateUser(user)

	router := setupTestRouter(userHandler)

	// Обновляем имя
	updatePayload := `{"first_name": "UpdatedName"}`

	req, _ := http.NewRequest("PUT", "/users/"+createdUser.ID.String(), strings.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "UpdatedName")
}

/*func TestUpdateUserWithEmptyFields(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	user := models.User{
		FirstName: "Ivan", LastName: "Petrov",
		Email: "ivan.petrov@example.com", Password: "securepass", Role: "user",
	}
	createdUser, _ := userService.CreateUser(user)

	payload := `{"first_name": "", "email": ""}`

	req, _ := http.NewRequest("PUT", "/users/"+createdUser.ID.String(), strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid first name")
}*/

func TestUpdateUserWithInvalidEmail(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	user := models.User{
		FirstName: "Lena", LastName: "Ivanova",
		Email: "lena.ivanova@example.com", Password: "pass123", Role: "user",
	}
	createdUser, _ := userService.CreateUser(user)

	payload := `{"email": "invalid-email"}`

	req, _ := http.NewRequest("PUT", "/users/"+createdUser.ID.String(), strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email")
}

func TestUpdateUser_InvalidJSON(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	// Создаем пользователя
	user := models.User{
		FirstName: "Ivan", LastName: "Petrov",
		Email: "ivan.petrov@example.com", Password: "secure123", Role: "user",
	}
	createdUser, _ := userService.CreateUser(user)

	// Невалидный JSON (пропущена кавычка)
	invalidJSON := `{"first_name": "Invalid}`

	req, _ := http.NewRequest("PUT", "/users/"+createdUser.ID.String(), strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
}


func TestDeleteUserHandler(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)

	user := models.User{
		FirstName: "Bob",
		LastName:  "Marley",
		Email:     "bob.marley@example.com",
		Password:  "password123",
		Role:      "user",
	}
	createdUser, _ := userService.CreateUser(user)

	router := setupTestRouter(userHandler)

	req, _ := http.NewRequest("DELETE", "/users/"+createdUser.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Проверим, что пользователь действительно удален
	_, err := userService.GetUser(createdUser.ID)
	assert.Error(t, err)
}

func TestDeleteNonExistingUser(t *testing.T) {
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	router := setupTestRouter(userHandler)

	nonExistingID := uuid.New()

	req, _ := http.NewRequest("DELETE", "/users/"+nonExistingID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
