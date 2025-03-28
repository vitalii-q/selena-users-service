package utils

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Фиксированная соль для тестов
const fixedSalt = "somefixedsalt"

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func HashPasswordForTests(password string) (string, error) {
    // Здесь можно использовать фиксированную соль для тестов.
    // В реальной ситуации соль должна быть уникальной, но для тестов это фиксированная строка.
    saltedPassword := fmt.Sprintf("%s%s", fixedSalt, password)
    
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    
    return string(hashedPassword), nil
}