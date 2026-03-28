package config

import (
	"log"
	"os"
)

// Env содержит все настройки приложения
type Env struct {
	Env           string
	Port          string
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBPort        string
	DBSSLMode     string
	ProjectSuffix string
}

// LoadEnv загружает конфиг из env переменных и проверяет обязательные
func LoadEnv() *Env {
	env := &Env{
		Env:           os.Getenv("APP_ENV"),
		Port:          os.Getenv("USERS_SERVICE_PORT"),
		DBHost:        os.Getenv("USERS_POSTGRES_DB_HOST"),
		DBUser:        os.Getenv("USERS_POSTGRES_DB_USER"),
		DBPassword:    os.Getenv("USERS_POSTGRES_DB_PASS"),
		DBName:        os.Getenv("USERS_POSTGRES_DB_NAME"),
		DBPort:        os.Getenv("USERS_POSTGRES_DB_PORT_INNER"),
		DBSSLMode:     os.Getenv("USERS_POSTGRES_DB_SSLMODE"),
	}

	// Проверка обязательных переменных
	if env.Port == "" {
		log.Fatal("USERS_SERVICE_PORT is not set")
	}
	if env.DBHost == "" || env.DBUser == "" || env.DBPassword == "" || env.DBName == "" || env.DBPort == "" {
		log.Fatal("One or more required database environment variables are missing")
	}

	// SSLMode по умолчанию
	if env.DBSSLMode == "" {
		if env.ProjectSuffix == "prod" {
			env.DBSSLMode = "require"
		} else {
			env.DBSSLMode = "disable"
		}
	}

	return env
}