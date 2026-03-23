package main

import (
	"context"
	"log"
	"time"

	"github.com/vitalii-q/selena-users-service/internal/database"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/router"
	"github.com/vitalii-q/selena-users-service/internal/server"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
	//"github.com/vitalii-q/selena-users-service/internal/logger"
)

func main() {
	// --- Context with cancel for graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Logger setup ---
	//logger.Setup()

	// --- Database connection ---
	dbPool, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbPool.Close()

	// --- Services and handlers ---
	passwordHasher := &utils.BcryptHasher{}
	hotelClient := external_services.NewHotelServiceClient()

	userService := services.NewUserService(dbPool, passwordHasher, hotelClient)
	userHandler := handlers.NewUserHandler(userService, hotelClient)
	authService := services.NewAuthService(dbPool)
	authHandler := &handlers.OAuthHandler{
		UserService: userService,
		AuthService: authService,
	}
	userHotelsHandler := handlers.NewUserHotelsHandler(hotelClient)
	locationsHandler := handlers.NewLocationsHandler(hotelClient)

	// --- Router setup ---
	r := router.SetupRouter(dbPool, userHandler, authHandler, userHotelsHandler, locationsHandler)

	// --- HTTP server ---
	srv := server.StartServer(r)

	// graceful shutdown
	server.GracefulShutdown(srv, cancel, 5*time.Second)
}