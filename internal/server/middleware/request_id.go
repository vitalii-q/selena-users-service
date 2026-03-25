package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID adds X-Request-ID header to every request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {

		requestID := c.GetHeader("X-Request-ID")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		// store request id in context
		c.Set("request_id", requestID)

		// return header to client
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}