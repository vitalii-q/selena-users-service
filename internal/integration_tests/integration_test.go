package integration_tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	//"path/filepath"
	//"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
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
func setupTestContainer() (container testcontainers.Container, dbPool *pgxpool.Pool, err error) {
	ctx := context.Background()

	// Запуск контейнера
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
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
		return nil, nil, err
	}

	// Получаем порт
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Формируем URL
	dbURL := fmt.Sprintf("postgres://test_user:test_password@%s:%s/test_db?sslmode=disable", host, port.Port())
	//logrus.Infof("Подключение к БД testcontainers: %s", dbURL)

	// Создаём pool
	dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, nil, err
	}

	// Проверка подключения
	if err := testDBconnection(ctx, dbPool); err != nil {
		return nil, nil, err
	}

	// Применяем миграции
	err = applyMigrations(ctx, host, port.Port())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return container, dbPool, nil
}

func testDBconnection(ctx context.Context, dbPool *pgxpool.Pool) error {
	var version string
	err := dbPool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return fmt.Errorf("Ошибка подключения к БД: %w", err)
	}
	logrus.Infof("Успешное подключение к контейнеру PostgreSQL. Версия: %s", version)
	return nil
}

// Функция для применения миграций через shell-скрипт
func applyMigrations(ctx context.Context, host, port string) error {
	//logrus.Infof("Host: %s", host)
	//logrus.Infof("Port: %s", port)

    // Извлекаем параметры подключения
    dbUser := "test_user"
    dbPassword := "test_password"
    dbName := "test_db"
	dbHost := host
    dbPort := port

	// Получаем root директорию
	rootDir, err := utils.GetRootDir()
	if err != nil {
		log.Fatal(err)
	}
	scriptPath := filepath.Join(rootDir, "db", "migrate_test.sh")
	migrationsDir := filepath.Join(rootDir, "db", "migrations")
	logrus.Infof("Root directory: %s", rootDir)

	// Запускаем shell-скрипт для применения миграций
	cmd := exec.Command(scriptPath, dbUser, dbPassword, dbHost, dbPort, dbName, migrationsDir, rootDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

// Основная функция для запуска тестов
func TestMain(m *testing.M) {
	container, dbPool, err := setupTestContainer()
	if err != nil {
		log.Fatalf("failed to setup test container: %v", err)
	}
	if dbPool == nil {
		log.Fatal("dbPool is nil, can't proceed with tests")
	}
	//logrus.Infof("dbPool: %#v", dbPool)
	defer container.Terminate(context.Background())
	defer dbPool.Close()
	//logrus.Infof("dbPool2: %#v", dbPool)

	// Запуск тестов
	os.Exit(m.Run())
}

// Интеграционный тест для CreateUser
func TestCreateUser(t *testing.T) {
	logrus.Infof("dbPool3: %#v", dbPool)

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

