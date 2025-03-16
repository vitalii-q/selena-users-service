package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Настройка логирования
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Получаем порт из переменной окружения (по умолчанию 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Инициализация Gin
	r := gin.Default()

	// Простой эндпоинт
	r.GET("/", func(c *gin.Context) {
		logrus.Info("GET / hit")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, users-service!"})
	})

	// Запуск сервера
	logrus.Infof("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		logrus.Fatalf("Server failed to start: %v", err)
	}
}
