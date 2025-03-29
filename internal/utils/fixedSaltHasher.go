package utils

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

// Фиксированная соль для тестов
const fixedSalt = "somefixedsalt"

// FixedSaltHasher — структура для хеширования с фиксированной солью
type FixedSaltHasher struct{}

// HashPassword — метод хеширования пароля с фиксированной солью
func (h *FixedSaltHasher) HashPassword(password string) (string, error) {
    saltedPassword := fmt.Sprintf("%s%s", fixedSalt, password)
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedPassword), nil
}

// CheckPasswordHash проверяет, соответствует ли пароль хэшу
func (h *FixedSaltHasher) CheckPasswordHash(password, hash string) bool {
    saltedPassword := fmt.Sprintf("%s%s", fixedSalt, password)
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword)) == nil
}

