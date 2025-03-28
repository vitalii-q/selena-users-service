package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

// MockUserService - это мок для интерфейса UserServiceInterface
type MockUserService struct {
	mock.Mock
}

// Метод для создания пользователя
func (m *MockUserService) CreateUser(user models.User) (models.User, error) {
	args := m.Called(user)
	return args.Get(0).(models.User), args.Error(1)
}

// Метод для получения пользователя
func (m *MockUserService) GetUser(id uuid.UUID) (models.User, error) {
	args := m.Called(id)
	return args.Get(0).(models.User), args.Error(1)
}

// Метод для обновления пользователя
func (m *MockUserService) UpdateUser(id uuid.UUID, user models.User) (models.User, error) {
	args := m.Called(id, user)
	return args.Get(0).(models.User), args.Error(1)
}

// Метод для удаления пользователя
func (m *MockUserService) DeleteUser(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func setupRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/users", handler.CreateUserHandler)
	return r
}

func TestCreateUserHandler(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	userService := services.NewUserService(mockDB)
	userHandler := NewUserHandler(userService)

	router := setupRouter(userHandler)

	// Пароль для теста
	plainPassword := "hashedpassword"

	// Хешируем пароль в тесте (как это делает сервис)
	hashedPassword, err := utils.HashPasswordForTests(plainPassword)
	assert.NoError(t, err)

	newUser := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		Password:  plainPassword, // Используем обычный пароль
		Role:      "user",
	}

	// Генерируем новый ID
	userID := uuid.New()
	// Ожидаем, что будет передано в запрос
	mockDB.ExpectQuery(`INSERT INTO users`).
		WithArgs("John", "Doe", "johndoe@example.com", hashedPassword, "user").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(userID, time.Now(), time.Now()))

	// Тело запроса
	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Log("Response Body:", w.Body.String()) // вывод тела запроса

	// Проверяем, что статус 201 Created
	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.ExpectationsWereMet()
}

func TestCreateUserHandler_InvalidJSON(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	userService := services.NewUserService(mockDB)
	userHandler := NewUserHandler(userService)
	router := setupRouter(userHandler)

	// Некорректный JSON (например, просто строка)
	body := []byte(`invalid json`)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
}

func TestCreateUserHandler_MissingField(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	userService := services.NewUserService(mockDB)
	userHandler := NewUserHandler(userService)
	router := setupRouter(userHandler)

	// JSON без email
	newUser := map[string]string{
		"first_name": "John",
		"last_name":  "Doe",
		"password":   "hashedpassword",
		"role":       "user",
	}

	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
}

func TestCreateUserHandler_DBError(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	userService := services.NewUserService(mockDB)
	userHandler := NewUserHandler(userService)
	router := setupRouter(userHandler)

	// Данные пользователя
	newUser := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johndoe@example.com",
		Password:  "hashedpassword",
		Role:      "user",
	}

	// Эмуляция ошибки при вставке в БД (например, дубликат email)
	mockDB.ExpectQuery(`INSERT INTO users`).
		WithArgs("John", "Doe", "johndoe@example.com", "", "user").
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "duplicate key value violates unique constraint")

	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockUserService)
	handler := &UserHandler{service: mockService}
	router.GET("/users/:id", handler.GetUserHandler)

	userID := uuid.New()
	expectedUser := models.User{ID: userID, FirstName: "John", LastName: "Doe"}
	mockService.On("GetUser", userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("GET", "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetUserHandler_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := &UserHandler{}
	router.GET("/users/:id", handler.GetUserHandler)

	req, _ := http.NewRequest("GET", "/users/invalidUUID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserHandler_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockUserService)
	handler := &UserHandler{service: mockService}
	router.GET("/users/:id", handler.GetUserHandler)

	userID := uuid.New()
	mockService.On("GetUser", userID).Return(models.User{}, errors.New("user not found"))

	req, _ := http.NewRequest("GET", "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockUserService)
	handler := &UserHandler{service: mockService}
	router.PUT("/users/:id", handler.UpdateUserHandler)

	userID := uuid.New()
	updatedUser := models.User{ID: userID, FirstName: "Updated", LastName: "User"}
	mockService.On("UpdateUser", userID, updatedUser).Return(updatedUser, nil)

	body, _ := json.Marshal(updatedUser)
	req, _ := http.NewRequest("PUT", "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateUserHandler_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := &UserHandler{}
	router.PUT("/users/:id", handler.UpdateUserHandler)

	req, _ := http.NewRequest("PUT", "/users/invalidUUID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUserHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := &UserHandler{}
	router.PUT("/users/:id", handler.UpdateUserHandler)

	req, _ := http.NewRequest("PUT", "/users/"+uuid.New().String(), bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUserHandler_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockUserService)
	handler := &UserHandler{service: mockService}
	router.PUT("/users/:id", handler.UpdateUserHandler)

	userID := uuid.New()
	mockService.On("UpdateUser", userID, mock.Anything).Return(models.User{}, errors.New("user not found"))

	body, _ := json.Marshal(models.User{ID: userID})
	req, _ := http.NewRequest("PUT", "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockUserService)
	handler := &UserHandler{service: mockService}
	router.DELETE("/users/:id", handler.DeleteUserHandler)

	userID := uuid.New()
	mockService.On("DeleteUser", userID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteUserHandler_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := &UserHandler{}
	router.DELETE("/users/:id", handler.DeleteUserHandler)

	req, _ := http.NewRequest("DELETE", "/users/invalidUUID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserHandler_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockUserService)
	handler := &UserHandler{service: mockService}
	router.DELETE("/users/:id", handler.DeleteUserHandler)

	userID := uuid.New()
	mockService.On("DeleteUser", userID).Return(errors.New("user not found"))

	req, _ := http.NewRequest("DELETE", "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
