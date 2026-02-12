# Тестовое задание на позицию backend разработчика

Результат выполнения задания нужно будет оформить здесь же, на гитхабе. Желательный срок выполнения 1 день с момента получения задания.
В качестве ответа __не нужно__ присылать никаких ZIP-архивов и наборов файлов. Все ваши ответы должны быть оформлены на https://github.com/.

## Задание

- Необходимо сделать небольшой web server
- __Обязательно__: postgres, без ORM, golang, оформить README.md, где будут указаны необходимые ENV переменные и инструкция к запуску
- Дополнения в виде документации, валидации и прочего - не обязательно, но будет плюсом

## Функционал

__endpoint 1:__

- Нужно обратиться к списку предметов skinport (https://docs.skinport.com/#items)
и отобразить массив объектов, где помимо прочего должны быть указаны две минимальные цены на предмет (одна цена — tradable, другая — нет)
параметры app_id и currency - default, базу здесь использовать не нужно, в эндпоинте необходимо использовать кеширование

__endpoint 2:__

- Есть табличка users с полями id и balance, там есть один юзер с id = 1
- Необходимо реализовать списание баланса пользователя
- Необходимо образовать историю списаний с баланса в виде - было-стало, когда, в кошельке какого user'a

__Пример__: юзер покупает какой-то предмет на сайте за $100. Баланс не должен быть ниже нуля.

---

## Реализация

### Технологический стек

- **Язык:** Go 1.25.7
- **База данных:** PostgreSQL 16 (без ORM, чистый SQL)
- **External API:** Skinport API v1
- **Архитектура:** Clean Architecture (handlers --> services --> repository)

### Структура проекта

```
.
├── cmd/server/          # Точка входа приложения
├── internal/
│   ├── config/         # Управление конфигурацией
│   ├── handlers/       # HTTP обработчики запросов
│   ├── models/         # Модели данных
│   ├── repository/     # Слой работы с БД (raw SQL)
│   └── services/       # Бизнес-логика
├── pkg/
│   ├── cache/          # In-memory кэш
│   ├── database/       # Подключение к БД
│   ├── errors/         # Пользовательские ошибки
│   ├── middlewares/    # HTTP middleware (gzip-сжатие)
│   ├── ratelimiter/    # Rate limiting для внешних API
│   └── skinport/       # Клиент Skinport API
└── migrations/         # SQL схема базы данных
```

### Быстрый старт

#### Требования

- Go 1.21 или выше
- Docker и Docker Compose

#### Установка и запуск

1. **Клонирование и настройка:**
   ```bash
   git clone git@github.com:Imomali1/backend-test-golang.git
   cd backend-test-golang
   cp .env.example .env
   # Отредактируйте .env при необходимости (стандартные значения работают с Docker)
   ```

2. **Запуск приложения:**
   ```bash
   docker-compose up -d
   ```
   
Сервер запустится на http://localhost:8080

#### Альтернатива: Использование Makefile

```bash
make run      # Запустить приложения
make stop     # Остановить приложения
make clean    # Удалить все контейнеры и volumes
```

#### Развертывание с Docker (Production)

Для полного развертывания всего стека (база данных + приложение) в контейнерах:

```bash
# Запустить все сервисы (PostgreSQL + приложение)
docker-compose up -d

# Проверить статус
docker-compose ps

# Посмотреть логи приложения
docker-compose logs -f app

# Остановить все сервисы
docker-compose down

# Пересобрать образ после изменений кода
docker-compose up -d --build
```

**Примечание:** При использовании `docker-compose up -d` приложение будет доступно на http://localhost:8080, а PostgreSQL на localhost:5432 (для внешних подключений).

**Multi-stage Dockerfile:**
- **Stage 1 (builder):** Сборка Go-бинарника с отключенным CGO для статической линковки
- **Stage 2 (runtime):** Минимальный Alpine-образ (без build-инструментов)
- **Security:** Запуск от non-root пользователя, health check
- **Размер итогового образа:** ~15-20 MB (вместо ~300 MB с full Go image)

### Переменные окружения

| Переменная | Обязательно | По умолчанию | Описание                                     |
|------------|-------------|--------------|----------------------------------------------|
| `ADDR` | Да | `:8080`      | Адрес и порт сервера                         |
| `DB_URL` | Да | -            | PostgreSQL connection string                 |
| `SKINPORT_ADDR` | Да | -            | Базовый URL Skinport API                     |
| `SKINPORT_CLIENT_ID` | Нет | -            | Client ID для Skinport API (опционально)     |
| `SKINPORT_CLIENT_SECRET` | Нет | -            | Client Secret для Skinport API (опционально) |
| `CACHE_TTL` | Нет | `300`        | Время жизни кэша в секундах                  |
| `CACHE_CLEANUP_INTERVAL` | Нет | `60`         | Интервал очистки кэша в секундах             |

**Примечание:** Skinport API работает без авторизации, но с более строгими rate limits. С авторизацией лимит выше.

### API Endpoints

#### 1. GET /api/v1/items

Получение списка предметов из Skinport API с двумя минимальными ценами (tradable/non-tradable).

**Особенности:**
- ✅ Кэширование ответов (по умолчанию 5 минут)
- ✅ Автоматическое gzip-сжатие
- ✅ Обработка ошибок внешнего API
- ✅ Rate limiting

**Примечание:** По умолчанию допускается 8 запросов за 5 минут (rate limiting).
Если получен HTTP-код 429 от API, проверяется заголовок Retry-After.
При наличии валидного заголовка запросы блокируются до указанного времени,
в противном случае — на 5 минут по умолчанию.


**Пример запроса:**
```bash
curl http://localhost:8080/api/v1/items
```

**Пример ответа:**
```json
{
  "success": true,
  "payload": [
    {
      "market_hash_name": "AK-47 | Redline (Field-Tested)",
      "currency": "EUR",
      "min_price_tradable": 25.99,
      "min_price_non_tradable": 23.50
    },
    {
      "market_hash_name": "AWP | Asiimov (Field-Tested)",
      "currency": "EUR",
      "min_price_tradable": 85.00,
      "min_price_non_tradable": 82.00
    }
  ]
}
```

#### 2. POST /api/v1/withdraw

Списание баланса пользователя с сохранением истории транзакций.

**Headers (обязательно):**
- `Content-Type: application/json`
- `X-Idempotency-Key: <UUID>` - уникальный ключ для предотвращения дублей

**Request Body:**
```json
{
  "user_id": 1,
  "amount": 10.50
}
```

**Особенности:**
- ✅ Идемпотентность (повторный запрос с тем же ключом безопасен)
- ✅ Атомарность (используются транзакции и блокировки записи на уровне бд)
- ✅ Валидация (баланс не может стать отрицательным)
- ✅ История транзакций (было-стало, время, пользователь)

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: $(uuidgen)" \
  -d '{"user_id": 1, "amount": 10.50}'
```

**Пример ответа:**
```json
{
  "success": true,
  "payload": {
    "id": 1,
    "idempotency_key": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": 1,
    "balance_before": "100.00",
    "balance_after": "89.50",
    "amount": "10.50",
    "created_at": "2026-02-01T10:30:00Z"
  }
}
```

**Идемпотентность:** Повторная отправка с тем же `X-Idempotency-Key` вернет оригинальную транзакцию без создания дубликата.

#### 3. GET /api/v1/user/balance

Получение текущего баланса пользователя.

**Query Parameters:**
- `user_id` (обязательно) - ID пользователя

**Пример:**
```bash
curl "http://localhost:8080/api/v1/user/balance?user_id=1"
```

**Ответ:**
```json
{
  "success": true,
  "balance": "89.50",
  "user_id": 1
}
```

#### 4. GET /api/v1/user/transactions

Получение истории транзакций пользователя.

**Query Parameters:**
- `user_id` (обязательно) - ID пользователя

**Пример:**
```bash
curl "http://localhost:8080/api/v1/user/transactions?user_id=1"
```

**Ответ:**
```json
{
  "success": true,
  "payload": [
    {
      "id": 1,
      "idempotency_key": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": 1,
      "balance_before": "100.00",
      "balance_after": "89.50",
      "amount": "10.50",
      "created_at": "2024-02-11T10:30:00Z"
    }
  ]
}
```

#### 5. GET /health

Health check endpoint для мониторинга.

```bash
curl http://localhost:8080/health
# Ответ: OK
```

### Схема базы данных

```sql
-- Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица транзакций (история: было-стало-когда)
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    idempotency_key VARCHAR UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    balance_before NUMERIC(15, 2) NOT NULL,  -- было
    balance_after NUMERIC(15, 2) NOT NULL,   -- стало
    amount NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()  -- когда
);

INSERT INTO users (id, balance) VALUES (1, 1000.00);
```

Схема автоматически создается при запуске `docker-compose up -d`.

### Архитектурные решения

#### Почему Raw SQL?

- ✅ Полный контроль над запросами
- ✅ Лучшая производительность для простых операций
- ✅ Явное управление транзакциями
- ✅ Четкое понимание происходящего в БД

#### Реализация идемпотентности

- Заголовок `X-Idempotency-Key` в формате UUID (обязателен)
- UNIQUE constraint на `idempotency_key` в БД предотвращает дубли
- При повторе возвращается оригинальная транзакция
- Следует best practices HTTP идемпотентности

#### Стратегия кэширования

- In-memory кэш (подходит для single-instance приложения)
- Cache-aside паттерн: проверка кэша → запрос к API → сохранение в кэш
- Настраиваемый TTL через переменную окружения
- **Для production:** рекомендуется Redis для distributed caching, иначе при каждом запуске кэш очищается

#### Обработка ошибок

- Кастомные типы ошибок
- Структурированные ответы с корректными HTTP статусами
- Graceful degradation при сбоях внешних API
- Автоматический rollback транзакций при ошибках БД

### Тестирование

#### Ручное тестирование

```bash
# 1. Health check
curl http://localhost:8080/health

# 2. Получить предметы (может занять несколько секунд при первом запросе)
curl http://localhost:8080/api/v1/items

# 3. Списать баланс
curl -X POST http://localhost:8080/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: $(uuidgen)" \
  -d '{"user_id": 1, "amount": 100.00}'

# 4. Проверить новый баланс
curl "http://localhost:8080/api/v1/user/balance?user_id=1"

# 5. Посмотреть историю транзакций
curl "http://localhost:8080/api/v1/user/transactions?user_id=1"

# 6. Тест идемпотентности (используем тот же ключ повторно)
curl -X POST http://localhost:8080/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{"user_id": 1, "amount": 50.00}'
# Повторный запрос вернет ту же транзакцию, баланс не изменится

# 7. Тест валидации: попытка списать больше, чем есть на балансе
curl -X POST http://localhost:8080/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: $(uuidgen)" \
  -d '{"user_id": 1, "amount": 999999.00}'
# Должна вернуться ошибка: insufficient balance
```

#### Обработанные edge cases

- ✅ Отрицательные суммы списания → Ошибка валидации
- ✅ Списание больше баланса → Insufficient balance error
- ✅ Дублирующиеся idempotency keys → Возврат оригинальной транзакции
- ✅ Несуществующий user ID → User not found error
- ✅ Отсутствие idempotency key → Bad request error
- ✅ Rate limit Skinport API → Возврат 429 с retry-after
- ✅ Сбои транзакций БД → Автоматический rollback

### Технические детали

#### Зависимости

```txt
github.com/andybalholm/brotli   // Brotli декомпрессия для Skinport API
github.com/google/uuid          // Валидация UUID
github.com/joho/godotenv        // Загрузка .env файлов
github.com/lib/pq               // PostgreSQL драйвер
github.com/shopspring/decimal   // Точная арифметика для денежных сумм
```

#### Организация кода

- **Clean Architecture:** Четкое разделение слоев handlers, services, repository
- **Dependency Injection:** Зависимости передаются через конструкторы
- **Error Wrapping:** Использование `fmt.Errorf` с `%w` для цепочки ошибок
- **Type Safety:** `decimal.Decimal` для денег (избегаем проблем с float точностью)
- **Concurrency:** Thread-safe кэш с mutex блокировками

### Что готово для production

- ✅ Database transactions для консистентности 
- ✅ Идемпотентность для безопасных retry 
- ✅ Rate limiting для внешних API 
- ✅ Структурированная обработка ошибок 
- ✅ Connection pooling 
- ✅ Gzip сжатие 
- ✅ Валидация входных данных 
- ✅ Использование decimal для денег

### Что добавить для production

- Structured logging (zerolog, zap)
- Metrics & monitoring (Prometheus)
- Distributed tracing (Jaeger, OpenTelemetry)
- Redis cache вместо in-memory
- Graceful shutdown handling
- Unit & integration tests
- CI/CD pipeline
- API documentation (OpenAPI/Swagger)
- Authentication & authorization
- Per-user rate limiting
- Database migrations tool (golang-migrate)
