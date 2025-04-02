package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	//"io"

	//"io"

	"os"
	//"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"

	//"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	//"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

var dbPool *pgxpool.Pool
var postgresContainer testcontainers.Container

// Функция для загрузки переменных окружения
func loadEnv() {
	// Загружаем переменные окружения, которые могут быть использованы в тестах
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "disable")
}

// Функция для запуска PostgreSQL контейнера
func startPostgresContainer() (testcontainers.Container, error) {
	// Запуск контейнера PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13", // Можно выбрать другую версию PostgreSQL
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432"), // Ожидаем, что порт 5432 откроется
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "../../db/migrations", // Путь на хосте с миграциями
				ContainerFilePath: "/migrations",          // Путь в контейнере
			},
			{
				HostFilePath:      "../../scripts/migrate_test.sh", // Путь на хосте с миграциями
				ContainerFilePath: "/scripts/migrate_test.sh",        // Путь в контейнере
			},
		},
	}

	// Создаем и запускаем контейнер
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	// Получаем проброшенный порт для подключения
	port, err := container.MappedPort(context.Background(), "5432")
	if err != nil {
		return nil, err
	}

	// Устанавливаем переменные окружения для подключения
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", port.Port())

	return container, nil
}

func applyMigrations(container testcontainers.Container) error {
	// Получаем параметры для подключения к базе данных
	port, err := container.MappedPort(context.Background(), "5432")
	if err != nil {
		return fmt.Errorf("не удалось получить порт контейнера: %v", err)
	}

	// Формируем строку подключения
	databaseUrl := fmt.Sprintf("postgres://%s:%s@localhost:%s/testdb?sslmode=disable",
		"test_user", "test_password", port.Port())

	// Открываем соединение с базой данных
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return fmt.Errorf("не удалось открыть соединение с БД: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("не удалось создать драйвер для БД: %v", err)
	}

	// Применяем миграции
	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations", // Путь внутри контейнера
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать миграции с экземпляром базы данных: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %v", err)
	}

	return nil
}

func TestMain(m *testing.M) {
	// Загружаем переменные окружения
	loadEnv()

	// Запускаем PostgreSQL контейнер
	container, err := startPostgresContainer()
	if err != nil {
		panic(fmt.Sprintf("Ошибка при запуске контейнера PostgreSQL: %v", err))
	}

	// Применяем миграции
	if err := applyMigrations(container); err != nil {
		panic(fmt.Sprintf("Ошибка применения миграций: %v", err))
	}

	// Выполнение тестов
	code := m.Run()

	// Остановка контейнера после выполнения тестов
	err = container.Terminate(context.Background())
	if err != nil {
		fmt.Printf("Ошибка при остановке контейнера: %v", err)
	}

	os.Exit(code)
}


// Интеграционный тест для CreateUser
func TestCreateUser(t *testing.T) {
	// Создаем объект passwordHasher (можно использовать реальную реализацию)
	passwordHasher := &utils.BcryptHasher{}
	userService := services.NewUserServiceImpl(dbPool, passwordHasher)

	// Создаем нового пользователя
	user := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		Role:      "user",
	}

	// Выполняем создание пользователя через сервис
	createdUser, err := userService.CreateUser(user)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)

	// Проверяем, что пользователь был создан
	assert.NotNil(t, createdUser.ID)
	assert.Equal(t, user.FirstName, createdUser.FirstName)
	assert.Equal(t, user.LastName, createdUser.LastName)
	assert.Equal(t, user.Email, createdUser.Email)
	assert.Equal(t, user.Role, createdUser.Role)
}
