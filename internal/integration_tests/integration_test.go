package integration_tests

import (
	"context"
	"fmt"
	"os"
	//"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services"
	"github.com/vitalii-q/selena-users-service/internal/utils"
)

var dbPool *pgxpool.Pool
//var postgresContainer testcontainers.Container

// Функция для загрузки переменных окружения
func loadEnv(host, port string) {
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port)
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "disable")
}

// Функция для запуска PostgreSQL контейнера
func startPostgresContainer(ctx context.Context) (testcontainers.Container, string, string, error) {
	// создаем и запускаем контейнер
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:13", // Можно выбрать другую версию PostgreSQL
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test_user",
				"POSTGRES_PASSWORD": "test_password",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForListeningPort("5432"), // Ожидаем, что порт 5432 откроется
		},
		Started: true,
		/*Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "../../db/migrations", // Путь на хосте с миграциями
				ContainerFilePath: "/migrations",          // Путь в контейнере
			},
			{
				HostFilePath:      "../../db/migrate_test.sh", // Путь на хосте с миграциями
				ContainerFilePath: "/db/migrate_test.sh",        // Путь в контейнере
			},
		},*/
	})

	if err != nil {
		return nil, "", "", err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", "", err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", "", err
	}

	return container, host, port.Port(), nil
}

// Функция для применения миграций через golang-migrate (не работает)
/*func applyMigrations(container testcontainers.Container) error {
    // Логируем рабочую директорию
    cwd, err := os.Getwd()
    if err != nil {logrus.Fatalf("Ошибка получения текущей рабочей директории: %v", err)}
    logrus.Infof("Текущая рабочая директория: %s", cwd)

	// Проверим наличие одного из файлов миграции
	projectRoot := filepath.Join(cwd, "..", "..")
	migrationPath := filepath.Join(projectRoot, "db", "migrations", "V1__create_users_table.up.sql")
	
	if _, err := os.Stat(migrationPath); err == nil {
		logrus.Infof("Файл миграции найден: %s", migrationPath)
	} else {
		logrus.Errorf("Файл миграции не найден: %s", migrationPath)
	}


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
	migrationsPath := filepath.Join(projectRoot, "db", "migrations")
	absMigrationsPath, err := filepath.Abs(migrationsPath) // absolute path to migrations
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь: %v", err)
	}

	sourceURL := fmt.Sprintf("file://%s", absMigrationsPath)
	logrus.Infof("Абсолютный путь к миграциям (sourceURL): %s", sourceURL)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать миграции с экземпляром базы данных: %v", err)
	}

	entries, err := os.ReadDir(absMigrationsPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать директорию миграций: %v", err)
	}

	for _, entry := range entries {
		logrus.Infof("Файл в директории миграций: %s", entry.Name())
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %v", err)
	}

	return nil
}*/

// Функция для применения миграций через shell-скрипт
func applyMigrations(ctx context.Context, db *pgxpool.Pool, migrationsPath string) error {
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return err
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		queries := strings.Split(string(content), ";")
		for _, query := range queries {
			q := strings.TrimSpace(query)
			if q != "" {
				_, err := db.Exec(ctx, q)
				if err != nil {
					return fmt.Errorf("failed to execute query in %s: %w", file, err)
				}
			}
		}
	}
	return nil
}

// Основная функция для запуска тестов
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Запускаем PostgreSQL контейнер
	container, host, port, err := startPostgresContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("Ошибка при запуске контейнера PostgreSQL: %v", err))
	}
	defer container.Terminate(ctx)

	// Загружаем переменные окружения на основе контейнера
	loadEnv(host, port)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		panic(fmt.Sprintf("Не удалось подключиться к базе данных: %v", err))
	}
	defer dbPool.Close()

	// Применяем миграции
	if err := applyMigrations(ctx, dbPool, "../../database/migrations"); err != nil {
		panic(fmt.Sprintf("Ошибка применения миграций: %v", err))
	}

	// Выполнение тестов
	code := m.Run()

	// Завершаем выполнение
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