# Документация проекта CNPF Feeder Backend

Добро пожаловать в документацию проекта! Здесь собрана вся информация о проекте.

## 📚 Навигация по документации

### 🏗️ Архитектура

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Описание архитектуры проекта (Onion Architecture)
- **[ONION_ARCHITECTURE_STATUS.md](./ONION_ARCHITECTURE_STATUS.md)** - Статус миграции на Onion Architecture
- **[ONION_ARCHITECTURE_MIGRATION.md](./ONION_ARCHITECTURE_MIGRATION.md)** - Инструкции по миграции
- **[MIGRATION_PROGRESS.md](./MIGRATION_PROGRESS.md)** - Прогресс миграции методов

### 🐳 Docker

- **[README_DOCKER.md](./README_DOCKER.md)** - Основная документация по Docker
- **[QUICK_START_DOCKER.md](./QUICK_START_DOCKER.md)** - Быстрый старт через Docker
- **[DOCKER_COMMANDS.md](./DOCKER_COMMANDS.md)** - Все команды для работы с Docker
- **[FIX_DOCKER.md](./FIX_DOCKER.md)** - Решение проблем с Docker

### 🚀 Разработка и настройка

- **[QUICK_TEST.md](./QUICK_TEST.md)** - Быстрое тестирование API
- **[GRAPHQL_EXAMPLES.md](./GRAPHQL_EXAMPLES.md)** - Примеры GraphQL запросов
- **[MONGODB_SETUP.md](./MONGODB_SETUP.md)** - Настройка MongoDB
- **[CHECK_DATABASE.md](./CHECK_DATABASE.md)** - Проверка базы данных
- **[SECURITY.md](./SECURITY.md)** - Безопасность и HTTPS
- **[FRONTEND_BACKEND_INTEGRATION.md](./FRONTEND_BACKEND_INTEGRATION.md)** - Интеграция с Frontend

---

## 🚀 Быстрый старт

### Локальная разработка

```bash
# Установить зависимости
go mod download

# Сгенерировать GraphQL код
go generate ./graph

# Запустить сервер
go run ./cmd/graph
```

### Docker

```bash
# Запустить через скрипт
./START_BACKEND.sh

# Или вручную
docker compose up mongo backend
```

Подробнее: [QUICK_START_DOCKER.md](./QUICK_START_DOCKER.md)

---

## 📖 Основные документы

Для начала работы рекомендуется прочитать:
1. [ARCHITECTURE.md](./ARCHITECTURE.md) - Понимание архитектуры
2. [ONION_ARCHITECTURE_STATUS.md](./ONION_ARCHITECTURE_STATUS.md) - Статус Onion Architecture
3. [README_DOCKER.md](./README_DOCKER.md) - Работа с Docker

---

## 📊 Статус проекта

### ✅ Готово - Onion Architecture полностью реализована!
- ✅ Onion Architecture структура создана и работает
- ✅ Доменные сущности созданы и расширены всеми полями
- ✅ Интерфейсы репозиториев определены и используются
- ✅ **Все методы UseCase реализованы** (Auth, User, Reports, Competitions, Admin)
- ✅ **Все resolver'ы используют UseCase** (не репозитории напрямую)
- ✅ **Репозитории удалены из Resolver**
- ✅ Конвертеры entity → GraphQL в UseCase
- ✅ Backend компилируется и запускается
- ✅ **100% соответствие Onion Architecture**

Подробнее: [MIGRATION_PROGRESS.md](./MIGRATION_PROGRESS.md), [ONION_ARCHITECTURE_STATUS.md](./ONION_ARCHITECTURE_STATUS.md)

---

## 🔗 Полезные ссылки

- GraphQL Playground: `http://localhost:4000/` (в debug режиме)
- Health Check: `http://localhost:4000/health`
- GraphQL Endpoint: `http://localhost:4000/graphql`
