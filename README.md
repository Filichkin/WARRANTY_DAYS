# WARRANTY_DAYS

Сервис на Go для работы с заявками по VIN и расчета дней ремонта в рамках текущего гарантийного года.

## Что делает проект

- Возвращает список заявок по VIN (регистронезависимо).
- Считает количество дней ремонта в рамках текущего гарантийного года.
- Поддерживает JWT-аутентификацию (`register/login/refresh`).
- Ограничивает доступ к бизнес-эндпоинтам только для авторизованных пользователей.

## Технологический стек

- Go `1.25`
- HTTP: стандартный `net/http`
- База данных: PostgreSQL
- ORM: `gorm` + `gorm.io/driver/postgres`
- Конфиг: переменные окружения (`.env`), загрузка через `godotenv`
- Логирование: `log/slog` (dev: text, prod: json)
- JWT: `github.com/golang-jwt/jwt/v5`

## Структура проекта

- `cmd/api/main.go` — точка входа, DI и запуск HTTP-сервера.
- `internal/config` — загрузка и валидация конфигурации.
- `internal/db` — подключение и настройки пула БД.
- `internal/models` — модели (`Claim`, `User`).
- `internal/repo` — доступ к данным (`ClaimRepo`, `UserRepo`).
- `internal/service` — бизнес-логика авторизации (`AuthService`).
- `internal/auth` — генерация и валидация JWT.
- `internal/httpapi/handler` — HTTP-хендлеры.
- `internal/httpapi/middleware` — middleware (auth + request logging).
- `internal/httpapi/router` — маршрутизация.
- `migrations` — SQL-миграции.

## Конфигурация

Используются переменные окружения:

- `APP_ENV`
- `LOG_LEVEL`
- `HTTP_ADDR` (по умолчанию `:8080`)
- `DB_HOST` (по умолчанию `127.0.0.1`)
- `DB_PORT` (по умолчанию `5432`)
- `DB_USER` (обязательный)
- `DB_PASSWORD` (обязательный)
- `DB_NAME` (обязательный)
- `DB_SSLMODE` (по умолчанию `disable`)
- `JWT_SECRET` (обязательный, минимум 32 символа)
- `JWT_ISSUER` (по умолчанию `warranty_days`)
- `JWT_ACCESS_TTL` (по умолчанию `15m`)
- `JWT_REFRESH_TTL` (по умолчанию `168h`)

## Запуск

1. Поднять PostgreSQL и создать БД (например, `warranty_days`).
2. Применить миграции:

```bash
psql "postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable" -f migrations/001_create_claims.sql
psql "postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable" -f migrations/002_create_users.sql
```

3. Заполнить `.env`.

### Dev (с autoreload)

1. Установить `air`:

```bash
go install github.com/air-verse/air@latest
```

2. Убедиться, что `$(go env GOPATH)/bin` есть в `PATH` (если `air` не находится).
3. Запускать сервис:

```bash
air
```

`air` использует конфиг из `.air.toml` и автоматически пересобирает/перезапускает сервер при изменении файлов.

### Dev (без autoreload, через `go run`)

Можно запускать и так:

```bash
go run ./cmd/api
```

Важно: при изменениях в коде процесс не перезапускается автоматически, нужно остановить и запустить команду заново.

### Prod (без autoreload)

Собрать и запустить бинарник:

```bash
go build -o bin/api ./cmd/api
./bin/api
```

## Auth (JWT)

### Публичные эндпоинты

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `GET /health`

### Защищенные эндпоинты

- `GET /claims?vin=...`
- `GET /claims/warranty-year?vin=...`

Для защищенных эндпоинтов нужен заголовок:

```text
Authorization: Bearer <access_token>
```

### Формат токенов

`/auth/login` и `/auth/refresh` возвращают:

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer"
}
```

## API

### Проверка доступности

- `GET /health`
- Ответ: `ok`

### Получить заявки по VIN

- `GET /claims?vin=XXX`
- Ответ: JSON-массив `Claim`.

### Рассчитать warranty-year repair days

- `GET /claims/warranty-year?vin=XXX`
- Ответ:

```json
{
  "items": [
    {
      "claim": {
        "id": 1,
        "vin": "XXX",
        "retail_date": "2024-01-10T00:00:00Z",
        "ro_open_date": "2025-10-10T00:00:00Z",
        "ro_close_date": "2025-10-25T00:00:00Z"
      },
      "repair_days": 11
    }
  ],
  "retail_date": "2024-01-10T00:00:00Z",
  "warranty_period": {
    "start": "2025-01-10T00:00:00Z",
    "end": "2026-01-09T00:00:00Z"
  },
  "total_days": 11
}
```

## Инструкция для Postman

1. Создай коллекцию и переменную `baseUrl = http://localhost:8080`.
2. Выполни `POST {{baseUrl}}/auth/register`.

Body (`raw`, `JSON`):

```json
{
  "email": "test@example.com",
  "password": "StrongPass123"
}
```

Ожидаемо: `201 Created`.

3. Выполни `POST {{baseUrl}}/auth/login`.

Body:

```json
{
  "email": "test@example.com",
  "password": "StrongPass123"
}
```

Ожидаемо: `200 OK` и токены.

4. Сохрани `access_token` в переменную Postman `accessToken`, `refresh_token` в `refreshToken`.
5. Для запроса `GET {{baseUrl}}/claims/warranty-year?vin=XWENE81BBM0000385` добавь Authorization:

- Type: `Bearer Token`
- Token: `{{accessToken}}`

6. Проверка refresh: `POST {{baseUrl}}/auth/refresh`.

Body:

```json
{
  "refresh_token": "{{refreshToken}}"
}
```

Ожидаемо: новая пара токенов.

7. Негативные проверки:

- Без токена на `/claims` -> `401`.
- Неверный пароль на `/auth/login` -> `401`.
- Повторная регистрация того же email -> `409`.

## Линтинг и форматирование

```bash
gofmt -w $(rg --files -g '*.go')
golangci-lint run ./...
```
