.PHONY: help setup run stop clean

help:
	@echo "Доступные команды:"
	@echo "  make setup    - Запустить PostgreSQL"
	@echo "  make run      - Запустить сервер"
	@echo "  make stop     - Остановить PostgreSQL"
	@echo "  make clean    - Удалить все контейнеры"

setup:
	@echo "Запуск PostgreSQL..."
	docker-compose up -d
	@echo "Ждем готовности базы данных..."
	@sleep 5
	@echo "✓ База данных готова! Запустите: make run"

run:
	@echo "Запуск сервера на http://localhost:8080..."
	go run cmd/server/main.go

stop:
	@echo "Остановка PostgreSQL..."
	docker-compose stop
	@echo "✓ PostgreSQL остановлен"

clean:
	@echo "Удаление контейнеров и volumes..."
	docker-compose down -v
	@echo "✓ Все контейнеры удалены"
