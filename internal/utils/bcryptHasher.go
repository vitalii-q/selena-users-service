package utils

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

// BcryptHasher - структура для хеширования паролей через bcrypt
type BcryptHasher struct{}

// HashPassword - хеширует пароль
func (b *BcryptHasher) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePassword - сравнивает хеш и пароль
func (b *BcryptHasher) ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
