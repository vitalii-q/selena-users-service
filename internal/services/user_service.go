package services

import (
    "github.com/vitalii-q/selena-users-service/internal/models"
    "github.com/google/uuid"
)

// Интерфейс UserService
type UserServiceInterface interface {
    CreateUser(user models.User) (models.User, error)
    GetUser(id uuid.UUID) (models.User, error)
    UpdateUser(id uuid.UUID, updatedUser models.User) (models.User, error)
    DeleteUser(id uuid.UUID) error
}

// Конструктор для UserServiceImpl
func NewUserService(db db_interface) UserServiceInterface {
    return &UserServiceImpl{db: db}
}

