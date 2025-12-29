package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	//"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/vitalii-q/selena-users-service/internal/database"
	"github.com/vitalii-q/selena-users-service/internal/handlers"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

type RootResponse struct {
    Message string `json:"message"`
    Host    string `json:"host"`
}

func init() {
	setupLogger() // Настраиваем логирование
}

func main() {
	// Создаём контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Освобождаем ресурсы при выходе

	// Подключаемся к базе
	dbPool, err := pgxpool.New(ctx, database.GetDatabaseURL())
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

	hotelClient := external_services.NewHotelServiceClient()
	userHotelsHandler := handlers.NewUserHotelsHandler(hotelClient)

	// Запускаем сервер
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9065" // По умолчанию основной контейнер работает на 9065
	}
	r := setupRouter(userHandler, OAuthHandler, userHotelsHandler)

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

// setupRouter инициализирует маршрутизатор и эндпоинты
func setupRouter(
	userHandler *handlers.UserHandler, 
	authHandler *handlers.OAuthHandler,
	userHotelsHandler *handlers.UserHotelsHandler,
) *gin.Engine {
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

	// User CRUD
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

	// get all hotels from hotels-service (did it for basic communication between microservices)
	r.GET("/users/:id/hotels", userHotelsHandler.GetUserHotelsHandler)

	// API routers
	r.GET("/users", userHandler.GetUsersHandler) // get all users

	return r
}

// handleRoot отвечает на запросы к "/"
func handleRoot(c *gin.Context) {
    hostname, err := os.Hostname()
    if err != nil {
        hostname = "unknown"
    }

	publicIP := getPublicIPv4()

	logrus.Info("GET / hit")
	c.JSON(http.StatusOK, gin.H{
        "message":     "Hello, users-service!",
        "Private_DNS": hostname, // вернём hostname/имя инстанса
		"Public_IPv4": publicIP,
    })
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

// getPublicIPv4 — IMDSv2 поддержка для запроса публичного IPv4 EC2
func getPublicIPv4() string {
    client := &http.Client{}

    // 1. Получаем IMDSv2 token
    tokenReq, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
    if err != nil {
        log.Printf("IMDSv2 token request build failed: %v", err)
        return ""
    }
    tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

    tokenResp, err := client.Do(tokenReq)
    if err != nil {
        log.Printf("IMDSv2 token request failed: %v", err)
        return ""
    }
    defer tokenResp.Body.Close()

    token, err := ioutil.ReadAll(tokenResp.Body)
    if err != nil {
        log.Printf("IMDSv2 token read failed: %v", err)
        return ""
    }

    // 2. Делаем запрос public-ipv4 с токеном
    metaReq, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/public-ipv4", nil)
    if err != nil {
        log.Printf("Public IPv4 request build failed: %v", err)
        return ""
    }
    metaReq.Header.Set("X-aws-ec2-metadata-token", string(token))

    metaResp, err := client.Do(metaReq)
    if err != nil {
        log.Printf("Public IPv4 request failed: %v", err)
        return ""
    }
    defer metaResp.Body.Close()

    body, err := ioutil.ReadAll(metaResp.Body)
    if err != nil {
        log.Printf("Public IPv4 read failed: %v", err)
        return ""
    }

    return string(body)
}
