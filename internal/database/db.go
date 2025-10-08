package database

import (
	"fmt"
	"log"
	"os"
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
