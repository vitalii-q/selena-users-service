package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/config"
)

func main() {
	// Настройка логирования
	setupLogger()

	// Определение порта
	port := getPort()

	// Инициализация маршрутизатора
	r := setupRouter()

	// Запуск сервера
	logrus.Infof("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		logrus.Fatalf("Server failed to start: %v", err)
	}
}

// setupLogger настраивает логирование
func setupLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

// getPort получает порт из переменной окружения или использует значение по умолчанию
func getPort() string {
	// Получаем путь к файлу конфигурации из переменной окружения
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// Если переменная окружения не задана, используем путь по умолчанию
		configPath = "/config/config.yaml"
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}

	port := cfg.Server.Port
	return port
}

// setupRouter инициализирует маршрутизатор и эндпоинты
func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)

	return r
}

// handleRoot отвечает на запросы к "/"
func handleRoot(c *gin.Context) {
	logrus.Info("GET / hit")
	c.JSON(http.StatusOK, gin.H{"message": "Hello, users-service!"})
}

// handleHealth отвечает на запросы к "/health"
func handleHealth(c *gin.Context) {
	logrus.Info("Health check request")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
