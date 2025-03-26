package handlers

import (
	"net/http"
	//"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
)

// UserHandler - структура обработчика пользователей
type UserHandler struct {
	service *services.UserService
}

// NewUserHandler - конструктор UserHandler
func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// CreateUserHandler - обработчик для создания пользователя
func (h *UserHandler) CreateUserHandler(c *gin.Context) {
	var user models.User

	//logrus.Info("TEST 1")

	if err := c.ShouldBindJSON(&user); err != nil {
		logrus.Error("JSON binding error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	//logrus.Info("TEST 2")

	// Проверяем, все ли параметры переданы
	if user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Role == "" {
		c.JSON(400, gin.H{"error": "Invalid request"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updatedUser, err = h.service.UpdateUser(id, updatedUser)
	if err != nil {
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
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
