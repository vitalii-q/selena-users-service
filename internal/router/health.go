package router

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

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