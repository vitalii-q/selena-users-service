package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// manually assemble the connection string
func GetDatabaseURL() string {
	dbUser := os.Getenv("USERS_POSTGRES_DB_USER")
	dbPassword := os.Getenv("USERS_POSTGRES_DB_PASS")
	dbName := os.Getenv("USERS_POSTGRES_DB_NAME")
	dbHost := os.Getenv("USERS_POSTGRES_DB_HOST")
	dbPort := os.Getenv("USERS_POSTGRES_DB_PORT_INNER")

	if dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" {
		log.Fatal("One or more required database environment variables are missing")
	}

	sslmode := os.Getenv("USERS_POSTGRES_DB_SSLMODE")
	if sslmode == "" {
		if os.Getenv("PROJECT_SUFFIX") == "prod" {
			sslmode = "require"
		} else {
			sslmode = "disable"
		}
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslmode)
}

// Connect creates a pgx connection pool
func Connect(ctx context.Context) (*pgxpool.Pool, error) {

	databaseURL := GetDatabaseURL()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return pool, nil
}