package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	//"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/internal/dto"
	"github.com/vitalii-q/selena-users-service/internal/helpers"
	"github.com/vitalii-q/selena-users-service/internal/metrics"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
)

// UserHandler - обработчик HTTP-запросов, связанных с пользователями
type UserHandler struct {
	service            services.UserServiceInterface
	validator          *validator.Validate
	HotelServiceClient *external_services.HotelServiceClient
}

// NewUserHandler - конструктор UserHandler
func NewUserHandler(service services.UserServiceInterface, HotelServiceClient *external_services.HotelServiceClient) *UserHandler {
	return &UserHandler{
		service:            service,
		validator:          validator.New(),
		HotelServiceClient: HotelServiceClient,
	}
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func writeUserError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{Error: code, Message: message})
}

func isDuplicateEmailError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return strings.Contains(strings.ToLower(pgErr.ConstraintName), "email") || strings.Contains(strings.ToLower(pgErr.Detail), "email")
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") && (strings.Contains(message, "email") || strings.Contains(message, "users_email"))
}

// CreateUserHandler - обработчик для создания пользователя
func (h *UserHandler) CreateUserHandler(c *gin.Context) {
	var user models.User

	//logrus.Info("TEST 1")

	if err := c.ShouldBindJSON(&user); err != nil {
		logrus.WithError(err).Error("Invalid JSON request")
		writeUserError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request")
		return
	}

	// Логируем значение SSLMODE из env
	sslMode := os.Getenv("USERS_POSTGRES_DB_SSLMODE")
	logrus.Infof("SSL mode from env: %s", sslMode)

	//logrus.Info("TEST 2")

	// Валидация структуры
	if err := h.validator.Struct(user); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Validation failed", Message: "Invalid user data", Details: err.Error()})
		return
	}

	createdUser, err := h.service.CreateUser(user)
	if err != nil {
		if isDuplicateEmailError(err) {
			writeUserError(c, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "A user with this email already exists")
			return
		}
		logrus.WithError(err).Error("failed to create user")
		writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	metrics.UsersCreatedTotal.Inc() // Count only successfully created users.

	c.JSON(http.StatusCreated, dto.MapUser(createdUser))
}

// GetUserHandler - обработчик для получения пользователя по ID
func (h *UserHandler) GetUserHandler(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeUserError(c, http.StatusBadRequest, "INVALID_UUID", "Invalid UUID format")
		return
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		if err.Error() == "user not found" {
			writeUserError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		} else {
			logrus.WithError(err).Error("failed to get user")
			writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	// enrichment через LocationsClient
	if h.HotelServiceClient == nil {
		c.JSON(http.StatusOK, dto.MapUser(user))
		return
	}
	enrichedUsers, err := helpers.EnrichUsers(
		[]models.User{user},
		h.HotelServiceClient,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to enrich user with locations")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to enrich user with locations",
		})
		return
	}

	if len(enrichedUsers) == 0 {
		writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(http.StatusOK, enrichedUsers[0])
}

// UpdateUserHandler - обработчик для обновления данных пользователя
func (h *UserHandler) UpdateUserHandler(c *gin.Context) {
	idParam := c.Param("id")

	// Парсим UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeUserError(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user ID")
		return
	}

	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		logrus.WithError(err).Error("Invalid JSON request") // лог в консоль
		writeUserError(c, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON or wrong field types")
		return
	}

	if updatedUser.ID != uuid.Nil && updatedUser.ID != id {
		writeUserError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Cannot update ID")
		return
	}

	// Валидируем только те поля, которые присутствуют в запросе
	if updatedUser.FirstName != "" {
		if err := h.validator.Var(updatedUser.FirstName, "required"); err != nil {
			writeUserError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid first name")
			return
		}
	}

	if updatedUser.LastName != "" {
		if err := h.validator.Var(updatedUser.LastName, "required"); err != nil {
			writeUserError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid last name")
			return
		}
	}

	if updatedUser.Email != "" {
		if err := h.validator.Var(updatedUser.Email, "required,email"); err != nil {
			writeUserError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid email")
			return
		}
	}

	if updatedUser.Password != "" {
		if err := h.validator.Var(updatedUser.Password, "required,min=6"); err != nil {
			writeUserError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid password")
			return
		}
	}

	// Обновляем пользователя
	updatedUser, err = h.service.UpdateUser(id, updatedUser)
	if err != nil {
		if err.Error() == "user not found" {
			writeUserError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		if isDuplicateEmailError(err) {
			writeUserError(c, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "A user with this email already exists")
			return
		}
		logrus.WithError(err).Error("failed to update user")
		writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	c.JSON(http.StatusOK, dto.MapUser(updatedUser))
}

// DeleteUserHandler - обработчик для удаления пользователя
func (h *UserHandler) DeleteUserHandler(c *gin.Context) {
	idStr := c.Param("id")

	// Преобразуем строку в uuid.UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeUserError(c, http.StatusBadRequest, "INVALID_UUID", "Invalid UUID format")
		return
	}

	// Вызываем сервис для удаления пользователя
	err = h.service.DeleteUser(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "user not found" {
			writeUserError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		} else {
			logrus.WithError(err).Error("failed to delete user")
			writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUsersHandler — get a list of all users with country/city names
func (h *UserHandler) GetUsersHandler(c *gin.Context) {
	// read query parameter
	expand := c.Query("expand")

	if expand == "locations" {
		usersWithLocations, err := h.service.(*services.UserService).GetAllUsersWithLocations()
		if err != nil {
			logrus.WithError(err).Error("failed to get users with locations")
			writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"users": usersWithLocations,
			"count": len(usersWithLocations),
		})
		return
	}
	if paginatedService, ok := h.service.(services.PaginatedUserService); ok {
		page, size, sortBy, sortDirection, err := parsePagination(c)
		if err != nil {
			writeUserError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		users, total, err := paginatedService.GetUsersPage(page, size, sortBy, sortDirection)
		if err != nil {
			logrus.WithError(err).Error("failed to get paginated users")
			writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}
		content := make([]dto.UserResponse, 0, len(users))
		for _, user := range users {
			content = append(content, dto.MapUser(user))
		}
		totalPages := int((total + int64(size) - 1) / int64(size))
		c.JSON(http.StatusOK, gin.H{"content": content, "total_elements": total, "total_pages": totalPages})
		return
	}

	// Get users from service
	users, err := h.service.GetAllUsers()
	if err != nil {
		logrus.WithError(err).Error("failed to get users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch users",
		})
		return
	}

	// Return users
	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, dto.MapUser(user))
	}
	c.JSON(http.StatusOK, gin.H{"users": responses, "count": len(responses)})
}

func parsePagination(c *gin.Context) (int, int, string, string, error) {
	page, size := 0, 20
	var err error
	if value := c.Query("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil || page < 0 {
			return 0, 0, "", "", errors.New("page must be greater than or equal to zero")
		}
	}
	if value := c.Query("size"); value != "" {
		size, err = strconv.Atoi(value)
		if err != nil || size < 1 || size > 100 {
			return 0, 0, "", "", errors.New("size must be between 1 and 100")
		}
	}
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortDirection := c.DefaultQuery("sortDirection", "desc")
	allowed := map[string]bool{"id": true, "first_name": true, "last_name": true, "email": true, "role": true, "created_at": true, "updated_at": true}
	if !allowed[sortBy] {
		return 0, 0, "", "", errors.New("sortBy is not allowed")
	}
	if sortDirection != "asc" && sortDirection != "desc" {
		return 0, 0, "", "", errors.New("sortDirection must be asc or desc")
	}
	return page, size, sortBy, sortDirection, nil
}
