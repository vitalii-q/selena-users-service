package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitalii-q/selena-users-service/internal/metrics"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid self-observability noise
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		method := c.Request.Method
		path := c.FullPath()

		metrics.HTTPRequestsTotal.WithLabelValues(
			method,
			path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			method,
			path,
			status,
		).Observe(duration)
	}
}