package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vitalii-q/selena-users-service/internal/database"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/router"
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
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9065"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Starting users-service on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// --- Graceful shutdown ---
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down users-service...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server exited cleanly")
}