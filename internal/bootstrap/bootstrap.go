package bootstrap

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalii-q/selena-users-service/internal/database"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

// Bootstrap struct holds all services and handlers
type Bootstrap struct {
	DBPool             *pgxpool.Pool
	UserService        *services.UserService
	AuthService        *services.AuthService
	UserHandler        *handlers.UserHandler
	AuthHandler        *handlers.OAuthHandler
	UserHotelsHandler  *handlers.UserHotelsHandler
	LocationsHandler   *handlers.LocationsHandler
}

// NewBootstrap initializes all dependencies and returns Bootstrap struct
func NewBootstrap(ctx context.Context) *Bootstrap {
	// --- Database connection ---
	dbPool, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// --- Utilities ---
	passwordHasher := &utils.BcryptHasher{}

	// --- External services ---
	hotelClient := external_services.NewHotelServiceClient()

	// --- Services ---
	userService := services.NewUserService(dbPool, passwordHasher, hotelClient)
	authService := services.NewAuthService(dbPool)

	// --- Handlers ---
	userHandler := handlers.NewUserHandler(userService, hotelClient)
	authHandler := &handlers.OAuthHandler{
		UserService: userService,
		AuthService: authService,
	}
	userHotelsHandler := handlers.NewUserHotelsHandler(hotelClient)
	locationsHandler := handlers.NewLocationsHandler(hotelClient)

	return &Bootstrap{
		DBPool:            dbPool,
		UserService:       userService,
		AuthService:       authService,
		UserHandler:       userHandler,
		AuthHandler:       authHandler,
		UserHotelsHandler: userHotelsHandler,
		LocationsHandler:  locationsHandler,
	}
}