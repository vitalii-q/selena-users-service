# Указываем интерпретатор для Makefile
SHELL := /bin/bash

# Переменная для Go
GO := go

# Команда для тестов
test:
	$(GO) test ./... -v

# Команда для сборки
build:
	$(GO) build -o users-service .

# Команда для запуска сервиса
run:
	$(GO) run main.go

# Команда для форматирования кода
fmt:
	$(GO) fmt ./...

# Команда для очистки билдов
clean:
	rm -f users-service
