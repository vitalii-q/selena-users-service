package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs HTTP requests
func Logger() gin.HandlerFunc {
	skipPaths := map[string]struct{}{   // map is faster than []string
		"/health": {},
		"/ready":  {},
	}

	return func(c *gin.Context) {

		path := c.Request.URL.Path

		// skip logging for health endpoints
		if _, ok := skipPaths[path]; ok {
			c.Next()
			return
		}

		start := time.Now()

		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		requestID, _ := c.Get("request_id")

		log.Printf(
			"request_id=%v method=%s path=%s status=%d latency=%s",
			requestID,
			method,
			path,
			status,
			latency,
		)
	}
}