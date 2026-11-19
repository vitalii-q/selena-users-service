package middleware

import (
	"net/http"
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

		method := normalizedMethod(c.Request.Method)
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		metrics.HTTPRequestsTotal.WithLabelValues(
			method,
			route,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			method,
			route,
			status,
		).Observe(duration)
	}
}

func normalizedMethod(method string) string {
	switch method {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodConnect,
		http.MethodTrace:
		return method
	default:
		return "OTHER"
	}
}
