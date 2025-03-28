package utils

// PasswordHasher - интерфейс для хеширования паролей
type PasswordHasher interface {
	HashPassword(password string) (string, error)
}
