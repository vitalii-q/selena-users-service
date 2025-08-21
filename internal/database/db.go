package repository

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func ConnectDB(host, port, user, password, dbname string) (*pgx.Conn, error) {
	sslmode := os.Getenv("USERS_POSTGRES_DB_SSLMODE")
	if sslmode == "" {
		if os.Getenv("PROJECT_SUFFIX") == "prod" {
			sslmode = "require"
		} else {
			sslmode = "disable"
		}
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	log.Println("Connected to database successfully")
	return conn, nil
}
