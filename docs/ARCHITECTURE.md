# Архитектура проекта CNPF Feeder Backend

Проект построен по принципам **Onion Architecture** (Луковая архитектура), которая является разновидностью Clean Architecture.

## Структура проекта

```
cnpf-feeder-backend/
├── cmd/                    # Точки входа приложения
│   └── graph/             # GraphQL сервер
│       └── server.go
├── graph/                 # GraphQL слой (Presentation Layer)
│   ├── schema/           # GraphQL схемы (.graphql файлы)
│   ├── resolver/         # Реализация резолверов
│   ├── model/            # GraphQL модели (автогенерируются)
│   ├── generated/        # Сгенерированный код gqlgen
│   └── scalars/          # Кастомные скалярные типы
├── internal/              # Внутренняя логика приложения
│   ├── domain/           # Доменный слой (Domain Layer) - ЦЕНТР
│   │   ├── entity/       # Доменные сущности (не зависят ни от чего)
│   │   │   ├── user.go
│   │   │   ├── report.go
│   │   │   └── competition.go
│   │   └── config.go     # Конфигурация
│   ├── repository/       # Интерфейсы репозиториев (Domain Layer)
│   │   └── interface/   # Интерфейсы (определены в domain слое)
│   │       ├── user_repository.go
│   │       ├── report_repository.go
│   │       └── competition_repository.go
│   ├── usecase/          # Слой приложения (Application Layer)
│   │   ├── base.go       # Интерфейсы use cases
│   │   └── usecase.go    # Реализация use cases
│   ├── repository/       # Репозитории (Infrastructure Layer)
│   │   └── mongodb/      # MongoDB репозитории (реализуют интерфейсы)
│   │       ├── mongodb.go
│   │       ├── user_repository.go
│   │       ├── report_repository.go
│   │       └── competition_repository.go
│   ├── auth/             # Аутентификация
│   │   ├── jwt.go
│   │   ├── password.go
│   │   └── current_user.go
│   ├── errors/           # Обработка ошибок
│   │   └── translations.go
│   └── validation/       # Валидация
│       └── validation.go
├── gqlgen.yml            # Конфигурация gqlgen
└── go.mod                # Зависимости проекта
```

## Архитектурные слои (Onion Architecture)

### 1. Domain Layer (Центр) - `internal/domain/`

**Доменные сущности** (`internal/domain/entity/`):
- `User` - пользователь
- `Report` - отчет
- `Competition` - соревнование

**Характеристика**: Доменные сущности не зависят ни от чего - это чистый центр архитектуры.

**Интерфейсы репозиториев** (`internal/repository/interface/`):
- Определяют контракты для работы с данными
- Не зависят от конкретных реализаций (MongoDB, PostgreSQL и т.д.)

### 2. Application Layer - `internal/usecase/`

**UseCase интерфейс** (`internal/usecase/base.go`):
- Определяет все бизнес-операции
- Группирует методы по доменам:
  - Auth (Register, Login, Logout)
  - User (GetCurrentUser, UpdateProfile, UpdatePassword)
  - Reports (CRUD операции)
  - Competitions (CRUD операции)
  - Admin (управление пользователями)

**Реализация** (`internal/usecase/usecase.go`):
- Реализует интерфейс UseCase
- Использует интерфейсы репозиториев (не конкретные реализации)
- Работает с доменными сущностями
- Возвращает GraphQL модели для Presentation слоя
- **Статус**: Auth и User методы полностью реализованы, остальные требуют реализации

### 3. Infrastructure Layer - `internal/repository/mongodb/`

**MongoDB репозитории**:
- Реализуют интерфейсы из `internal/repository/interface/`
- Конвертируют между доменными сущностями и MongoDB документами
- Содержат детали работы с MongoDB

**Характеристика**: 
- Легко заменить на другую реализацию (PostgreSQL, in-memory и т.д.) без изменения UseCase
- Все методы работают с доменными сущностями
- Конвертируют между доменными сущностями и MongoDB документами

### 4. Presentation Layer - `graph/resolver/`

**Резолверы** (`graph/resolver/`):
- Реализуют GraphQL запросы
- Используют UseCase для бизнес-логики (Auth и User операции)
- Используют интерфейсы репозиториев для работы с данными (Reports, Competitions, Admin)
- Работают с доменными сущностями (entity.User, entity.Report, entity.Competition)
- Конвертируют доменные сущности в GraphQL модели через функции форматирования

**Конвертеры** (`graph/resolver/converters.go`):
- `formatUserFromEntity()` - конвертирует `*entity.User` → `*model.User`
- `formatReportFromEntity()` - конвертирует `*entity.Report` → `*model.Report`
- `formatCompetitionFromEntity()` - конвертирует `*entity.Competition` → `*model.Competition`

**Сервер** (`cmd/graph/server.go`):
- Инициализирует зависимости
- Настраивает маршрутизацию
- Связывает слои вместе

## Поток зависимостей (Onion Architecture)

```
┌─────────────────────────────────────────┐
│  Presentation (graph/resolver)          │
│  - Использует UseCase                   │
└─────────────────────────────────────────┘
           │ depends on
           ▼
┌─────────────────────────────────────────┐
│  Application (internal/usecase)         │
│  - Использует repository interfaces     │
│  - Работает с domain entities           │
└─────────────────────────────────────────┘
           │ depends on
           ▼
┌─────────────────────────────────────────┐
│  Domain (internal/domain/entity)        │
│  - Чистые сущности                     │
│  - НЕ зависит ни от чего               │
└─────────────────────────────────────────┘
           ▲ implements
           │
┌─────────────────────────────────────────┐
│  Infrastructure (internal/repository)    │
│  - Реализует интерфейсы                 │
│  - Конвертирует entity ↔ MongoDB         │
└─────────────────────────────────────────┘
```

**Принципы**:
1. ✅ Зависимости направлены внутрь (к центру - Domain)
2. ✅ Domain не зависит ни от чего
3. ✅ Infrastructure реализует интерфейсы из Domain
4. ✅ Application использует интерфейсы, а не конкретные реализации
5. ✅ Presentation использует Application, а не Infrastructure напрямую

## Потоки данных

### GraphQL запрос
```
Client → GraphQL Handler → Resolver → UseCase → Repository Interface → MongoDB Repository → Database
                                                                              ↓
Client ← GraphQL Response ← Resolver ← UseCase ← Repository Interface ← MongoDB Repository ← Database
```

### Пример: Регистрация пользователя
```
1. Client отправляет GraphQL mutation Register
2. Resolver.Register получает input
3. Resolver вызывает useCase.Register
4. UseCase.Register:
   - Валидирует данные
   - Проверяет существование пользователя через userRepo
   - Создает доменную сущность entity.User
   - Сохраняет через userRepo.Create
   - Генерирует JWT токен
   - Возвращает AuthResult
5. Resolver устанавливает cookie и возвращает результат
```

## Конфигурация

### Переменные окружения

**Основные**:
- `PORT` - порт GraphQL сервера (по умолчанию 4000)
- `GIN_MODE` - режим Gin (debug/release)
- `CORS_ORIGIN` - разрешенный origin для CORS

**База данных**:
- `MONGODB_URI` - URI подключения к MongoDB
- `MONGODB_NAME` - имя базы данных (опционально, извлекается из URI)

**Аутентификация**:
- `AUTH_SECRET` - секретный ключ для JWT токенов

**Логирование**:
- `LOGLEVEL` - уровень логирования (info, debug, warn, error)

## Основные домены

1. **Auth** - аутентификация и авторизация
2. **User** - управление пользователями
3. **Reports** - отчеты пользователей
4. **Competitions** - соревнования
5. **Admin** - административные функции

## Особенности реализации

1. **Onion Architecture** - четкое разделение на слои с правильным направлением зависимостей
2. **Repository Pattern** - абстракция доступа к данным через интерфейсы
3. **UseCase Pattern** - бизнес-логика изолирована от деталей реализации
4. **Dependency Injection** - зависимости передаются через конструкторы
5. **GraphQL First** - GraphQL как основной API интерфейс
6. **Domain Entities** - чистые доменные сущности в центре архитектуры

## Статус миграции

- ✅ **Структура Onion Architecture создана**
- ✅ **Доменные сущности созданы и расширены**
- ✅ **Интерфейсы репозиториев определены**
- ✅ **MongoDB репозитории реализуют интерфейсы**
- ✅ **UseCase использует интерфейсы репозиториев**
- ✅ **Все методы UseCase реализованы** (Auth, User, Reports, Competitions, Admin)
- ✅ **Все resolver'ы используют UseCase**
- ✅ **Репозитории удалены из Resolver**
- ✅ **Конвертеры entity → GraphQL в UseCase**
- ✅ **Backend компилируется и запускается**
- ✅ **100% соответствие Onion Architecture**

Подробнее: [MIGRATION_PROGRESS.md](./MIGRATION_PROGRESS.md), [ONION_ARCHITECTURE_STATUS.md](./ONION_ARCHITECTURE_STATUS.md)

## Рекомендации по развитию

1. ✅ Реализованы все методы UseCase (Auth, User, Reports, Competitions, Admin)
2. ✅ Все resolver'ы используют UseCase
3. ✅ Репозитории удалены из Resolver
4. ✅ Конвертеры entity → GraphQL в UseCase
5. ✅ **Onion Architecture полностью реализована**
6. Добавить unit тесты для UseCase и Repository
7. Добавить integration тесты для GraphQL API
8. Добавить метрики Prometheus
9. Добавить структурированное логирование (zap)
10. Добавить rate limiting
11. Реализовать graceful shutdown для всех компонентов
