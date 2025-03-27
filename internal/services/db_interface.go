package services

import (
	"context"
	"github.com/jackc/pgx/v5" // Используем типы из пакета pgx
	"github.com/jackc/pgx/v5/pgconn"
)

// Интерфейс для абстракции работы с базой данных
type db_interface interface {
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}
