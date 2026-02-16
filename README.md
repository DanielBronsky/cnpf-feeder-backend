# CNPF Feeder Backend (Go)

GraphQL API сервер для проекта CNPF Feeder на базе Go, gqlgen и Gin.

## Требования

- **Go >= 1.23**
- MongoDB (локально или через Docker)

## Установка

```bash
# Установить зависимости
go mod download

# Сгенерировать GraphQL код
go generate ./...

# Или вручную:
go run github.com/99designs/gqlgen generate
```

## Переменные окружения

Создайте файл `.env` в корне проекта (можно скопировать из `env.example`):

```env
MONGODB_URI=mongodb://localhost:27017/cnpf_feeder
AUTH_SECRET=change_this_to_a_long_random_string
PORT=4000
CORS_ORIGIN=http://localhost:3000
GIN_MODE=release
```

## Запуск

### Development (рекомендуется для разработки)

**Вариант 1: MongoDB в Docker, Backend локально** (быстрая разработка)

1. Запустите только MongoDB:
   ```bash
   docker compose -f docker-compose.dev.yml up mongo -d
   ```

2. Запустите Backend локально:
   ```bash
   go run ./cmd/graph
   ```

3. Откройте GraphQL Playground: `http://localhost:4000/`

**Вариант 2: Все локально**

Если MongoDB установлен локально:
```bash
go run ./cmd/graph
```

**Остановка MongoDB:**
```bash
docker compose -f docker-compose.dev.yml down
```

**Hot Reload (опционально):**

Для автоматической перезагрузки при изменении кода установите `air`:
```bash
go install github.com/cosmtrek/air@latest
```

Затем запускайте:
```bash
air
```

Сервер запустится на `http://localhost:4000/`

### Production

```bash
go build -o main ./cmd/graph
./main
```

## Docker

Подробная документация по Docker находится в папке `docs/`:
- `docs/README_DOCKER.md` - основная документация по Docker
- `docs/QUICK_START_DOCKER.md` - быстрый старт через Docker
- `docs/DOCKER_COMMANDS.md` - все команды для работы с Docker

## GraphQL Playground

В development режиме доступен GraphQL Playground по адресу `http://localhost:4000/`

## Docker

### Development окружение (только MongoDB)

Для локальной разработки используйте `docker-compose.dev.yml`:

```bash
# Запустить MongoDB
docker compose -f docker-compose.dev.yml up mongo -d

# Остановить MongoDB
docker compose -f docker-compose.dev.yml down

# Остановить и удалить данные
docker compose -f docker-compose.dev.yml down -v
```

### Production (только Backend)

```bash
docker build -t cnpf-feeder-backend .
docker run -p 4000:4000 --env-file .env cnpf-feeder-backend
```

### Полный стек (Frontend + Backend + MongoDB)

Используйте `docker-compose.yml` из Frontend проекта:

```bash
cd /Users/daniel/projects/cnpf.feeder.md
AUTH_SECRET="your_secret_here" docker compose up --build
```

Это поднимет:
- MongoDB на порту 27017
- Backend (GraphQL API) на порту 4000
- Frontend (Next.js) на порту 3000

## Структура проекта

```
.
├── cmd/
│   └── graph/
│       └── server.go         # Точка входа (GraphQL сервер)
├── graph/
│   ├── schema/               # GraphQL схемы
│   ├── resolver/            # Resolvers (реализация)
│   ├── model/               # Сгенерированные модели
│   ├── generated/           # Сгенерированный код (gqlgen)
│   └── scalars/             # Custom scalars (Date)
├── internal/
│   ├── domain/              # Конфигурация
│   ├── usecase/             # Бизнес-логика
│   ├── repository/          # Репозитории (MongoDB)
│   ├── auth/                # Аутентификация (JWT, password)
│   ├── errors/              # Обработка ошибок
│   └── validation/          # Валидация входных данных
├── docs/                    # Документация проекта
├── config/                  # Конфигурационные файлы
└── scripts/                 # Вспомогательные скрипты
├── go.mod
├── gqlgen.yml               # Конфигурация gqlgen
└── Dockerfile
```

## Генерация GraphQL кода

После изменения `graph/schema/schema.graphql` нужно сгенерировать код:

```bash
go generate ./graph
# или
cd graph && go run github.com/99designs/gqlgen@latest generate
```

## Документация

Вся документация проекта находится в папке `docs/`:

- **Архитектура**: `docs/ARCHITECTURE.md`
- **Docker**: `docs/README_DOCKER.md`, `docs/QUICK_START_DOCKER.md`
- **Миграция**: `docs/MIGRATION_COMPLETE.md`, `docs/RESTRUCTURING_SUMMARY.md`
- **Разработка**: `docs/DEVELOPMENT.md`, `docs/SETUP.md`

## API Endpoints

- `POST /graphql` - GraphQL endpoint
- `GET /health` - Health check endpoint
- `GET /` - GraphQL Playground (только в development)

## Особенности

- JWT аутентификация через httpOnly cookies
- Поддержка CORS для Frontend
- Scalar Date для работы с датами
- Валидация входных данных
- Graceful shutdown

## Миграция с TypeScript

Проект был переписан с TypeScript на Go. Старые файлы TypeScript находятся в папке `src/` (можно удалить после проверки).
