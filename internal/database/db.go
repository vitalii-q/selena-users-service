package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitalii-q/selena-users-service/internal/config"
)

// Manually assemble the connection string
func GetDatabaseURL(env *config.Env) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env.DBUser,
		env.DBPassword,
		env.DBHost,
		env.DBPort,
		env.DBName,
		env.DBSSLMode,
	)
}

// Connect creates a pgx connection pool
func Connect(ctx context.Context, env *config.Env) (*pgxpool.Pool, error) {
	log.Println("🗄️ Connecting to database...")

	databaseURL := GetDatabaseURL(env)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	log.Println("✅ Database connection established")
	return pool, nil
}