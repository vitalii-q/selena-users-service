package router

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter initializes Gin router with all routes and middleware
func SetupRouter(
	dbPool *pgxpool.Pool,
	userHandler *handlers.UserHandler,
	authHandler *handlers.OAuthHandler,
	userHotelsHandler *handlers.UserHotelsHandler,
	locationsHandler *handlers.LocationsHandler,
) *gin.Engine {
	r := gin.New()

	// --- Middleware ---
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/ready"},
	}))

	// --- Root & health checks ---
	r.GET("/", handleRoot)
	r.GET("/health", health)
	r.GET("/ready", ready(dbPool))
	r.GET("/protected", protected)

	// --- OAuth ---
	r.POST("/users/oauth2/authenticate", authHandler.Authenticate)

	// --- API routes ---
	api := r.Group("/api/v1")
	{
		api.POST("/users", userHandler.CreateUserHandler)
		api.GET("/users/:id", userHandler.GetUserHandler)
		api.PUT("/users/:id", userHandler.UpdateUserHandler)
		api.DELETE("/users/:id", userHandler.DeleteUserHandler)
		api.GET("/users", userHandler.GetUsersHandler)

		api.GET("/locations", locationsHandler.GetLocationsHandler)
	}

	// --- User Hotels ---
	r.GET("/users/:id/hotels", userHotelsHandler.GetUserHotelsHandler)

	return r
}

func handleRoot(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	publicIP := getPublicIPv4()

	logrus.Info("GET / hit")
	c.JSON(http.StatusOK, gin.H{
		"message":     "Hello, users-service!",
		"Private_DNS": hostname,
		"Public_IPv4": publicIP,
	})
}

