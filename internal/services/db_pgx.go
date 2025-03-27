package services

import (
	"context"

	"github.com/jackc/pgx/v5" // Используем pgx для CommandTag
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool" // Пул подключений pgx
)

// Реализация интерфейса Database с использованием pgxpool
type PgxDatabase struct {
	pool *pgxpool.Pool
}

// Новый конструктор для PgxDatabase
func NewPgxDatabase(pool *pgxpool.Pool) *PgxDatabase {
	return &PgxDatabase{pool: pool}
}

// Реализуем метод Exec
func (db *PgxDatabase) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...) // pgx.CommandTag используется здесь
}

// Реализуем метод QueryRow
func (db *PgxDatabase) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...) // Этот метод возвращает строку данных из базы
}
