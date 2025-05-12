package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// getRootDir ищет корневую директорию микросервиса (где находится go.mod)
func GetRootDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("не удалось получить текущую директорию: %w", err)
	}

	for {
		// Проверяем наличие файла go.mod
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Достигли корня файловой системы
			return "", fmt.Errorf("файл go.mod не найден — не удалось определить корень проекта")
		}
		dir = parent
	}
}
