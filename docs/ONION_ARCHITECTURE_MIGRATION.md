# Миграция на Onion Architecture - Инструкции

> ⚠️ **ВНИМАНИЕ**: Этот документ содержит исторические инструкции по миграции. **Миграция завершена полностью!**  
> Для актуального статуса см. [ONION_ARCHITECTURE_STATUS.md](./ONION_ARCHITECTURE_STATUS.md)

## ✅ Миграция завершена!

Все задачи миграции выполнены:

1. ✅ Созданы доменные сущности (`internal/domain/entity/`)
2. ✅ Созданы интерфейсы репозиториев (`internal/repository/interface/`)
3. ✅ MongoDB репозитории реализуют интерфейсы
4. ✅ UseCase использует интерфейсы репозиториев
5. ✅ **Все методы UseCase реализованы**
6. ✅ **Все resolver'ы используют UseCase**
7. ✅ **Репозитории удалены из Resolver**
8. ✅ `cmd/graph/server.go` правильно инициализирует зависимости

## 📝 Исторические инструкции (для справки)

### Что было сделано

#### 1. Реализованы все методы UseCase

Все методы в `internal/usecase/usecase.go` реализованы:
- Auth: `Register`, `Login`, `Logout`
- User: `GetCurrentUser`, `UpdateProfile`, `UpdatePassword`
- Reports: `GetReports`, `GetReport`, `CreateReport`, `UpdateReport`, `DeleteReport`
- Competitions: `GetCompetitions`, `GetCompetition`, `CreateCompetition`, `UpdateCompetition`, `DeleteCompetition`
- Admin: `GetAdminUsers`, `GetAdminUser`, `AdminUpdateUser`, `AdminDeleteUser`

#### 2. Мигрированы все Resolver'ы на UseCase

Все методы в `graph/resolver/schema.resolvers.go` используют `r.useCase` вместо прямых вызовов репозиториев.

**Пример миграции:**

**Было:**
```go
func (r *mutationResolver) CreateReport(ctx context.Context, input model.CreateReportInput) (*model.Report, error) {
    // ... много кода с прямыми вызовами r.reportRepo
    reportID, err := r.reportRepo.Create(ctx, reportEntity)
    // ...
}
```

**Стало:**
```go
func (r *mutationResolver) CreateReport(ctx context.Context, input model.CreateReportInput) (*model.Report, error) {
    user, err := getCurrentUserFromContext(ctx)
    if err != nil || user == nil {
        return nil, fmt.Errorf("Не авторизован")
    }
    
    // Конвертируем GraphQL uploads в UseCase PhotoUpload
    var photos []*usecase.PhotoUpload
    if input.Photos != nil {
        // ...
    }
    
    // Вызываем UseCase
    return r.useCase.CreateReport(ctx, user.ID, input.Title, input.Text, photos)
}
```

#### 3. Удалены репозитории из Resolver

**Было:**
```go
type Resolver struct {
    useCase usecase.UseCase
    userRepo        repository.UserRepository
    reportRepo      repository.ReportRepository
    competitionRepo repository.CompetitionRepository
}
```

**Стало:**
```go
type Resolver struct {
    useCase usecase.UseCase
}
```

#### 4. Обновлен NewResolver

**Было:**
```go
func NewResolver(useCase usecase.UseCase, userRepo repository.UserRepository, reportRepo repository.ReportRepository, competitionRepo repository.CompetitionRepository) *Resolver {
    return &Resolver{
        useCase:        useCase,
        userRepo:       userRepo,
        reportRepo:     reportRepo,
        competitionRepo: competitionRepo,
    }
}
```

**Стало:**
```go
func NewResolver(useCase usecase.UseCase) *Resolver {
    return &Resolver{
        useCase: useCase,
    }
}
```

#### 5. Обновлен server.go

**Было:**
```go
resolver := resolver.NewResolver(useCase, userRepo, reportRepo, competitionRepo)
```

**Стало:**
```go
resolver := resolver.NewResolver(useCase)
```

## 🔄 Правильный поток зависимостей (Onion Architecture)

```
┌─────────────────────────────────────────┐
│  Presentation Layer (graph/resolver)    │
│  - Использует UseCase                   │
│  - Конвертирует GraphQL → UseCase      │
└─────────────────────────────────────────┘
           │ depends on
           ▼
┌─────────────────────────────────────────┐
│  Application Layer (internal/usecase)    │
│  - Использует repository interfaces     │
│  - Работает с domain entities           │
│  - Возвращает GraphQL models            │
└─────────────────────────────────────────┘
           │ depends on
           ▼
┌─────────────────────────────────────────┐
│  Domain Layer (internal/domain/entity)  │
│  - Чистые сущности                      │
│  - НЕ зависит ни от чего                │
└─────────────────────────────────────────┘
           ▲ implements
           │
┌─────────────────────────────────────────┐
│  Infrastructure (internal/repository)   │
│  - Реализует repository interfaces      │
│  - Конвертирует entity ↔ MongoDB doc    │
└─────────────────────────────────────────┘
```

## ✅ Критерии завершения миграции

- [x] Все resolver'ы работают с entity типами
- [x] Созданы конвертеры entity ↔ GraphQL model
- [x] Resolver использует интерфейсы репозиториев
- [x] Backend компилируется без ошибок
- [x] Backend запускается и работает
- [x] **Все методы UseCase реализованы**
- [x] **Все методы Resolver используют `r.useCase`**
- [x] **Удалены временные репозитории из Resolver**

## 📚 Актуальная документация

- [ONION_ARCHITECTURE_STATUS.md](./ONION_ARCHITECTURE_STATUS.md) - текущий статус архитектуры
- [MIGRATION_PROGRESS.md](./MIGRATION_PROGRESS.md) - прогресс миграции
- [ARCHITECTURE.md](./ARCHITECTURE.md) - описание архитектуры
- [ONION_ARCHITECTURE_ANSWER.md](./ONION_ARCHITECTURE_ANSWER.md) - ответ на вопрос о соответствии Onion Architecture

## ✅ Вывод

**Миграция завершена успешно!**

Проект полностью соответствует Onion Architecture. Все слои правильно разделены, зависимости направлены внутрь к Domain, а Presentation использует только Application слой (UseCase).
