package router

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

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

// --- Handlers for root, health, readiness, protected ---

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

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "test ok"})
}

func ready(dbPool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := dbPool.Ping(ctx)
		if err != nil {
			logrus.Errorf("Readiness check failed: DB unreachable: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"db":     "down",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"db":     "up",
		})
	}
}

func protected(c *gin.Context) {
	logrus.Info("Protected check request")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// getPublicIPv4 — IMDSv2 support for EC2 public IPv4
func getPublicIPv4() string {
	client := &http.Client{}

	// 1. IMDSv2 token
	tokenReq, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		logrus.Errorf("IMDSv2 token request build failed: %v", err)
		return ""
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		logrus.Errorf("IMDSv2 token request failed: %v", err)
		return ""
	}
	defer tokenResp.Body.Close()

	token, err := json.Marshal(tokenResp.Body)
	if err != nil {
		logrus.Errorf("IMDSv2 token read failed: %v", err)
		return ""
	}

	// 2. Public IPv4 request
	metaReq, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/public-ipv4", nil)
	if err != nil {
		logrus.Errorf("Public IPv4 request build failed: %v", err)
		return ""
	}
	metaReq.Header.Set("X-aws-ec2-metadata-token", string(token))

	metaResp, err := client.Do(metaReq)
	if err != nil {
		logrus.Errorf("Public IPv4 request failed: %v", err)
		return ""
	}
	defer metaResp.Body.Close()

	body, err := json.Marshal(metaResp.Body)
	if err != nil {
		logrus.Errorf("Public IPv4 read failed: %v", err)
		return ""
	}

	return string(body)
}