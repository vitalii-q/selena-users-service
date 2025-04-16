package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword - проверка пароля
func CheckPassword(providedPassword, storedPasswordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(providedPassword))
	return err == nil
}
