package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	//"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
)

// UserHandler - обработчик HTTP-запросов, связанных с пользователями
type UserHandler struct {
	service   services.UserServiceInterface
	validator *validator.Validate
}

// NewUserHandler - конструктор UserHandler
func NewUserHandler(service services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: validator.New(),
	}
}

// CreateUserHandler - обработчик для создания пользователя
func (h *UserHandler) CreateUserHandler(c *gin.Context) {
	var user models.User

	//logrus.Info("TEST 1")

	if err := c.ShouldBindJSON(&user); err != nil {
		logrus.WithError(err).Error("Invalid JSON request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": "Invalid JSON request"})
		return
	}

	//logrus.Info("TEST 2")

	// Валидация структуры
	if err := h.validator.Struct(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	createdUser, err := h.service.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

// GetUserHandler - обработчик для получения пользователя по ID
func (h *UserHandler) GetUserHandler(c *gin.Context) {
	idStr := c.Param("id")

	// Преобразуем строку в uuid.UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUserHandler - обработчик для обновления данных пользователя
func (h *UserHandler) UpdateUserHandler(c *gin.Context) {
	idParam := c.Param("id")

	// Парсим UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		logrus.WithError(err).Error("Invalid JSON request") // лог в консоль
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": "Malformed JSON or wrong field types",
		})
		return
	}

	if updatedUser.ID != uuid.Nil && updatedUser.ID != id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update ID"})
		return
	}

	// Валидируем только те поля, которые присутствуют в запросе
	if updatedUser.FirstName != "" {
		if err := h.validator.Var(updatedUser.FirstName, "required"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid first name"})
			return
		}
	}

	if updatedUser.LastName != "" {
		if err := h.validator.Var(updatedUser.LastName, "required"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last name"})
			return
		}
	}

	if updatedUser.Email != "" {
		if err := h.validator.Var(updatedUser.Email, "required,email"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
			return
		}
	}

	if updatedUser.Password != "" {
		if err := h.validator.Var(updatedUser.Password, "required,min=6"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
			return
		}
	}

	// Обновляем пользователя
	updatedUser, err = h.service.UpdateUser(id, updatedUser)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}


// DeleteUserHandler - обработчик для удаления пользователя
func (h *UserHandler) DeleteUserHandler(c *gin.Context) {
	idStr := c.Param("id")

	// Преобразуем строку в uuid.UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Вызываем сервис для удаления пользователя
	err = h.service.DeleteUser(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
