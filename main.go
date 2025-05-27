package main

import (
	"context"
	"encoding/json"
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/config"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

func init() {
	setupLogger() // Настраиваем логирование
}

func main() {
	// Создаём контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Освобождаем ресурсы при выходе

	// Подключаемся к базе
	dbPool, err := pgxpool.New(ctx, getDatabaseURL())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close() // Закроется корректно при завершении программы

	// Создаём хешер паролей (обычный)
	passwordHasher := &utils.BcryptHasher{} 

	// Создаём сервис и обработчики
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)
	userHandler := handlers.NewUserHandler(userService)
	authService := services.NewAuthService(dbPool)

	// Создаём обработчик OAuth
	OAuthHandler := &handlers.OAuthHandler{
		UserService: userService,
		AuthService: authService,
	}

	// Запускаем сервер
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9065" // По умолчанию основной контейнер работает на 9065
	}
	r := setupRouter(userHandler, OAuthHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		logrus.Infof("Starting server on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed: %v", err)
		}
	}()

	// Ждем сигнала завершения (например, Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down server...")
	cancel()             // Завершаем контекст
	server.Shutdown(ctx) // Корректно останавливаем сервер
}

// setupLogger настраивает логирование
func setupLogger() {
	logrus.SetLevel(logrus.DebugLevel)         // Устанавливаем глобальный уровень логирования
	logrus.SetFormatter(&logrus.TextFormatter{ // Опционально: настраиваем формат логов
		FullTimestamp: true,
		ForceColors:   true,
	})
	logrus.SetOutput(os.Stdout)
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
func setupRouter(userHandler *handlers.UserHandler, authHandler *handlers.OAuthHandler) *gin.Engine {
	r := gin.Default()

	// Логгер для всех входящих запросов
	r.Use(func(c *gin.Context) {
		logrus.Infof("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// test routes
	r.GET("/", handleRoot)
	r.GET("/test", test)
	r.GET("/protected", protected)

	// Определяем маршруты
	r.POST("/users", userHandler.CreateUserHandler)
	r.GET("/users/:id", userHandler.GetUserHandler)
	r.PUT("/users/:id", userHandler.UpdateUserHandler)
	r.DELETE("/users/:id", userHandler.DeleteUserHandler)

	// authenticate
	r.POST("/users/oauth2/authenticate", authHandler.Authenticate)
	//r.GET("/oauth2/authorize", authHandler.GetAuthorize)
	//r.POST("/oauth2/token", authHandler.PostToken)

	b, _ := json.Marshal(authHandler) // +
	logrus.Debugf("authHandler: %s", string(b))

	//logrus.Debug("test!!!: s", authHandler)

	return r
}

// handleRoot отвечает на запросы к "/"
func handleRoot(c *gin.Context) {
	logrus.Info("GET / hit")
	c.JSON(http.StatusOK, gin.H{"message": "Hello, users-service!"})
}

// handleHealth отвечает на запросы к "/health"
func test(c *gin.Context) {
	logrus.Info("Test check request")
	c.JSON(http.StatusOK, gin.H{"status": "test ok"})
}

// protected отвечает на запросы к "/protected" защищен oauth2
func protected(c *gin.Context) {
	logrus.Info("Protected check request")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getDatabaseURL() string {
	// Собираем строку подключения вручную
	dbUser := os.Getenv("USERS_POSTGRES_DB_USER")
	dbPassword := os.Getenv("USERS_POSTGRES_DB_PASS")
	dbName := os.Getenv("USERS_POSTGRES_DB_NAME")
	dbHost := os.Getenv("USERS_POSTGRES_DB_HOST")
	dbPort := os.Getenv("USERS_POSTGRES_DB_PORT_INNER")

	if dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" {
		log.Fatal("One or more required database environment variables are missing (main.go)")
	}

	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	return databaseUrl
}
