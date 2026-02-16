# Интеграция Frontend и Backend

## Статус: ✅ Завершено

Все GraphQL запросы из Frontend проекта реализованы в Backend проекте.

## Реализованные запросы

### Queries (Запросы)

✅ **Me** - Получить текущего пользователя
- Реализовано в: `graph/resolver/schema.resolvers.go::Me()`

✅ **Reports** - Список отчетов с лимитом
- Реализовано в: `graph/resolver/schema.resolvers.go::Reports()`

✅ **Report** - Получить отчет по ID
- Реализовано в: `graph/resolver/schema.resolvers.go::Report()`

✅ **Competitions** - Список всех соревнований
- Реализовано в: `graph/resolver/schema.resolvers.go::Competitions()`

✅ **Competition** - Получить соревнование по ID
- Реализовано в: `graph/resolver/schema.resolvers.go::Competition()`

✅ **AdminUsers** - Список всех пользователей (только для админов)
- Реализовано в: `graph/resolver/schema.resolvers.go::AdminUsers()`

✅ **AdminUser** - Получить пользователя по ID (только для админов)
- Реализовано в: `graph/resolver/schema.resolvers.go::AdminUser()`

### Mutations (Мутации)

✅ **Register** - Регистрация нового пользователя
- Реализовано в: `graph/resolver/schema.resolvers.go::Register()`

✅ **Login** - Вход пользователя
- Реализовано в: `graph/resolver/schema.resolvers.go::Login()`

✅ **Logout** - Выход пользователя
- Реализовано в: `graph/resolver/schema.resolvers.go::Logout()`

✅ **UpdateProfile** - Обновление профиля пользователя
- Реализовано в: `graph/resolver/schema.resolvers.go::UpdateProfile()`
- Поддерживает загрузку аватара через GraphQL Upload

✅ **UpdatePassword** - Смена пароля
- Реализовано в: `graph/resolver/schema.resolvers.go::UpdatePassword()`

✅ **CreateReport** - Создание отчета
- Реализовано в: `graph/resolver/schema.resolvers.go::CreateReport()`
- Поддерживает загрузку фотографий через GraphQL Upload

✅ **UpdateReport** - Обновление отчета
- Реализовано в: `graph/resolver/schema.resolvers.go::UpdateReport()`
- Поддерживает добавление/удаление фотографий через GraphQL Upload

✅ **DeleteReport** - Удаление отчета
- Реализовано в: `graph/resolver/schema.resolvers.go::DeleteReport()`

✅ **CreateCompetition** - Создание соревнования
- Реализовано в: `graph/resolver/schema.resolvers.go::CreateCompetition()`

✅ **UpdateCompetition** - Обновление соревнования
- Реализовано в: `graph/resolver/schema.resolvers.go::UpdateCompetition()`

✅ **DeleteCompetition** - Удаление соревнования
- Реализовано в: `graph/resolver/schema.resolvers.go::DeleteCompetition()`

✅ **AdminUpdateUser** - Обновление пользователя (только для админов)
- Реализовано в: `graph/resolver/schema.resolvers.go::AdminUpdateUser()`

✅ **AdminDeleteUser** - Удаление пользователя (только для админов)
- Реализовано в: `graph/resolver/schema.resolvers.go::AdminDeleteUser()`

## Особенности реализации

### GraphQL Upload
- ✅ Поддержка загрузки файлов через GraphQL Upload scalar
- ✅ Валидация типов файлов (только изображения)
- ✅ Валидация размеров файлов (аватар: 512KB, фото: 2MB)

### Переводы ошибок
- ✅ Все ошибки переведены на русский язык
- ✅ Все ошибки начинаются с большой буквы
- ✅ Автоматический перевод MongoDB ошибок

### Аутентификация
- ✅ JWT токены в httpOnly cookies
- ✅ CORS настроен для Frontend
- ✅ Проверка прав доступа (admin/user)

## Docker Compose

### Запуск всего стека

```bash
cd /Users/daniel/projects/cnpf.feeder.md
docker compose up -d
```

Или через yarn:
```bash
yarn compose:up
```

### Сервисы

- **MongoDB**: порт 27017
- **Backend (GraphQL)**: порт 4000
- **Frontend (Next.js)**: порт 3000

### URL

- Frontend: http://localhost:3000
- Backend GraphQL: http://localhost:4000/graphql
- GraphQL Playground: http://localhost:4000/

## Структура проектов

```
/Users/daniel/projects/
├── cnpf.feeder.md/          # Frontend (Next.js)
│   ├── docker-compose.yml   # Docker Compose для всего стека
│   ├── src/
│   │   ├── lib/
│   │   │   └── graphql/     # GraphQL запросы и мутации
│   │   └── app/              # Next.js App Router
│   └── package.json
│
└── cnpf-feeder-backend/     # Backend (Go GraphQL)
    ├── graph/
    │   ├── schema/           # GraphQL схема
    │   ├── resolver/         # Резолверы
    │   └── scalars/          # Custom scalars
    ├── cmd/graph/            # Точка входа сервера
    ├── internal/             # Внутренние пакеты
    └── Dockerfile
```

## Проверка работы

1. **Запустить стек:**
   ```bash
   cd /Users/daniel/projects/cnpf.feeder.md
   docker compose up -d
   ```

2. **Проверить статус:**
   ```bash
   docker compose ps
   ```

3. **Открыть Frontend:**
   http://localhost:3000

4. **Открыть GraphQL Playground:**
   http://localhost:4000/

5. **Протестировать запрос:**
   ```graphql
   query {
     me {
       id
       email
       username
     }
   }
   ```

## Следующие шаги

Все запросы реализованы и протестированы. Проект готов к использованию!
