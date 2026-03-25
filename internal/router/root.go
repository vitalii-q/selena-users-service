package router

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func handleRoot(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	publicIP := getPublicIPv4()

	c.JSON(http.StatusOK, gin.H{
		"message":     "Hello, users-service!",
		"Private_DNS": hostname,
		"Public_IPv4": publicIP,
	})
}

