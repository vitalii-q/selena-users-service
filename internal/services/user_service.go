package services

import (
	"github.com/google/uuid"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

// UserService Interface
type UserServiceInterface interface {
	CreateUser(user models.User) (models.User, error)
	GetUser(id uuid.UUID) (models.User, error)
	UpdateUser(id uuid.UUID, updatedUser models.User) (models.User, error)
	DeleteUser(id uuid.UUID) error
	GetAllUsers() ([]models.User, error)

	HotelClient() *external_services.HotelServiceClient
}

// Constructor for UserServiceImpl
func NewUserService(db db_interface, passwordHasher utils.PasswordHasher) UserServiceInterface {
	if passwordHasher == nil {
		passwordHasher = &utils.BcryptHasher{} // Устанавливаем дефолтный хешер, если не передан
	}

	return &UserServiceImpl{
		db:             db,
		passwordHasher: passwordHasher,
	}
}
