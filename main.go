package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/config"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/services"
)

func main() {
	log.Fatal("test main.go")
	// Настройка логирования
	setupLogger()

	// Определение порта
	port := getPort()

	// Загружаем строку подключения из переменных окружения
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Подключаемся к базе через pgxpool
	dbPool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Создаём сервис пользователей
	userService := services.NewUserService(dbPool)

	// Создаём обработчик пользователей
	userHandler := handlers.NewUserHandler(userService)

	// Инициализация маршрутизатора
	r := setupRouter(userHandler)

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
func setupRouter(userHandler *handlers.UserHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)

	// Определяем маршруты
	r.POST("/users", userHandler.CreateUserHandler)
	r.GET("/users/:id", userHandler.GetUserHandler)
	r.PUT("/users/:id", userHandler.UpdateUserHandler)
	r.DELETE("/users/:id", userHandler.DeleteUserHandler)

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
