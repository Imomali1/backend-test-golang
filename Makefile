.PHONY: help setupdb runapp stopdb cleandb

help:
	@echo "Доступные команды:"
	@echo "  make setup    - Запустить PostgreSQL"
	@echo "  make run      - Запустить сервер"
	@echo "  make stop     - Остановить PostgreSQL"
	@echo "  make clean    - Удалить все контейнеры"

run:
	@echo "Запуск приложения"
	docker-compose up -d
	@echo "✓ База данных готова!"
	@echo "✓ Сервер запущен"

stop:
	@echo "Остановка приложений..."
	docker-compose stop
	@echo "✓ PostgreSQL остановлен"
	@echo "✓ Сервер остановлен"

clean:
	@echo "Удаление контейнеров и volumes..."
	docker-compose down -v
	@echo "✓ Все контейнеры удалены"

setupdb:
	@echo "Запуск PostgreSQL..."
	docker-compose up -d postgres
	@echo "Ждем готовности базы данных..."
	@sleep 5
	@echo "✓ База данных готова! Запустите: make runapp"

runapp:
	@echo "Запуск сервера на http://localhost:8080..."
	go run cmd/server/main.go

stopdb:
	@echo "Остановка PostgreSQL..."
	docker-compose stop postgres
	@echo "✓ PostgreSQL остановлен"

cleandb:
	@echo "Удаление контейнеров и volumes..."
	docker-compose down -v postgres
	@echo "✓ Все контейнеры удалены"
