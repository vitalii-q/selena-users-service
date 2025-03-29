package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/lib/pq"
)

var db *sql.DB

// Запускаем тестовый контейнер перед тестами
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Настройка контейнера с PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	// Создаем контейнер
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:         true,
	})
	if err != nil {
		log.Fatalf("Ошибка запуска контейнера: %s", err)
	}
	defer postgresC.Terminate(ctx)

	// Получаем IP и порт контейнера
	host, _ := postgresC.Host(ctx)
	port, _ := postgresC.MappedPort(ctx, "5432")

	// Подключаемся к PostgreSQL
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %s", err)
	}

	// Выполняем миграции (временно через SQL)
	_, err = db.Exec(`
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("Ошибка миграции: %s", err)
	}

	// Запускаем тесты
	code := m.Run()

	// Выход с кодом результата тестов
	os.Exit(code)
}

// Тест регистрации пользователя
func TestUserRegistration(t *testing.T) {
	// Вставляем пользователя в БД
	_, err := db.Exec(`INSERT INTO users (email, password) VALUES ($1, $2)`, "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Ошибка вставки пользователя: %s", err)
	}

	// Проверяем, что он существует
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE email=$1`, "test@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("Ошибка выборки пользователя: %s", err)
	}

	if count != 1 {
		t.Fatalf("Ожидали 1 пользователя, а получили %d", count)
	}
}
