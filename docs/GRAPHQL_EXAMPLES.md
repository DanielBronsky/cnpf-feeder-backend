# Примеры GraphQL запросов

Этот документ содержит примеры запросов для тестирования GraphQL API.

## 🎯 GraphQL Playground

Откройте в браузере: `http://localhost:4000/`

GraphQL Playground - это интерактивная среда для тестирования GraphQL запросов. Вы можете:
- Писать запросы в левой панели
- Нажимать кнопку "Play" для выполнения
- Видеть результаты в правой панели
- Просматривать документацию схемы (вкладка "Schema")

## 📝 Примеры запросов (Queries)

### 1. Health Check (через GraphQL)

```graphql
query {
  __typename
}
```

### 2. Получить текущего пользователя (Me)

**Требует авторизации** (нужен JWT токен в cookie или Authorization header)

```graphql
query {
  me {
    id
    email
    username
    isAdmin
    hasAvatar
    avatarUrl
  }
}
```

### 3. Получить список отчетов

```graphql
query {
  reports(limit: 10) {
    id
    title
    text
    createdAt
    author {
      id
      username
      hasAvatar
      avatarUrl
    }
    photos {
      url
    }
    canEdit
  }
}
```

### 4. Получить один отчет

```graphql
query {
  report(id: "REPORT_ID_HERE") {
    id
    title
    text
    createdAt
    updatedAt
    author {
      id
      username
    }
    photos {
      url
    }
    canEdit
  }
}
```

### 5. Получить список соревнований

```graphql
query {
  competitions {
    id
    title
    startDate
    endDate
    location
    individualFormat
    teamFormat
    fee
    teamLimit
    createdAt
  }
}
```

### 6. Получить одно соревнование

```graphql
query {
  competition(id: "COMPETITION_ID_HERE") {
    id
    title
    startDate
    endDate
    location
    tours {
      date
      time
    }
    openingDate
    openingTime
    individualFormat
    teamFormat
    fee
    teamLimit
    regulations
  }
}
```

### 7. Получить список пользователей (только для админов)

**Требует авторизации и прав администратора**

```graphql
query {
  adminUsers {
    id
    email
    username
    isAdmin
    hasAvatar
  }
}
```

## ✏️ Примеры мутаций (Mutations)

### 1. Регистрация пользователя

```graphql
mutation {
  register(input: {
    email: "test@example.com"
    username: "testuser"
    password: "password123"
    passwordConfirm: "password123"
  }) {
    ok
    token
  }
}
```

**С аватаром:**

```graphql
mutation {
  register(input: {
    email: "test@example.com"
    username: "testuser"
    password: "password123"
    passwordConfirm: "password123"
    avatar: null  # В Playground используйте вкладку "FILES" для загрузки
  }) {
    ok
    token
  }
}
```

### 2. Вход (Login)

```graphql
mutation {
  login(input: {
    login: "test@example.com"  # или username
    password: "password123"
  }) {
    ok
    token
  }
}
```

### 3. Выход (Logout)

**Требует авторизации**

```graphql
mutation {
  logout
}
```

### 4. Обновить профиль

**Требует авторизации**

```graphql
mutation {
  updateProfile(input: {
    username: "newusername"
  }) {
    id
    username
    email
    hasAvatar
    avatarUrl
  }
}
```

**Удалить аватар:**

```graphql
mutation {
  updateProfile(input: {
    removeAvatar: true
  }) {
    id
    hasAvatar
    avatarUrl
  }
}
```

### 5. Изменить пароль

**Требует авторизации**

```graphql
mutation {
  updatePassword(
    oldPassword: "oldpassword123"
    newPassword: "newpassword123"
  )
}
```

### 6. Создать отчет

**Требует авторизации**

```graphql
mutation {
  createReport(input: {
    title: "Мой первый отчет"
    text: "Текст отчета здесь..."
    photos: []  # В Playground используйте вкладку "FILES" для загрузки
  }) {
    id
    title
    text
    createdAt
    author {
      username
    }
  }
}
```

### 7. Обновить отчет

**Требует авторизации (только автор или админ)**

```graphql
mutation {
  updateReport(
    id: "REPORT_ID_HERE"
    input: {
      title: "Обновленный заголовок"
      text: "Обновленный текст"
    }
  ) {
    id
    title
    text
    updatedAt
  }
}
```

### 8. Удалить отчет

**Требует авторизации (только автор или админ)**

```graphql
mutation {
  deleteReport(id: "REPORT_ID_HERE")
}
```

### 9. Создать соревнование

**Требует авторизации**

```graphql
mutation {
  createCompetition(input: {
    title: "Чемпионат 2026"
    startDate: "2026-03-01"
    endDate: "2026-03-05"
    location: "Москва"
    individualFormat: true
    teamFormat: false
    tours: [
      { date: "2026-03-01", time: "10:00" }
      { date: "2026-03-02", time: "10:00" }
    ]
  }) {
    id
    title
    startDate
    endDate
  }
}
```

### 10. Обновить соревнование

**Требует авторизации**

```graphql
mutation {
  updateCompetition(
    id: "COMPETITION_ID_HERE"
    input: {
      title: "Обновленное название"
      location: "Санкт-Петербург"
    }
  ) {
    id
    title
    location
    updatedAt
  }
}
```

### 11. Удалить соревнование

**Требует авторизации**

```graphql
mutation {
  deleteCompetition(id: "COMPETITION_ID_HERE")
}
```

### 12. Обновить пользователя (админ)

**Требует авторизации и прав администратора**

```graphql
mutation {
  adminUpdateUser(
    id: "USER_ID_HERE"
    isAdmin: true
  ) {
    id
    username
    isAdmin
  }
}
```

### 13. Удалить пользователя (админ)

**Требует авторизации и прав администратора**

```graphql
mutation {
  adminDeleteUser(id: "USER_ID_HERE")
}
```

## 🔐 Авторизация

### Способ 1: Cookie (автоматически)

После выполнения `login` или `register`, токен сохраняется в cookie `cnpf_auth`. Все последующие запросы автоматически используют этот токен.

### Способ 2: Authorization Header

В GraphQL Playground:
1. Откройте вкладку "HTTP HEADERS" внизу
2. Добавьте:

```json
{
  "Authorization": "Bearer YOUR_JWT_TOKEN_HERE"
}
```

### Способ 3: Через cURL

```bash
# С cookie
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: cnpf_auth=YOUR_TOKEN" \
  -d '{"query":"{ me { id username } }"}'

# С Authorization header
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"query":"{ me { id username } }"}'
```

## 🧪 Тестирование через cURL

### Пример 1: Регистрация

```bash
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { register(input: { email: \"test@example.com\", username: \"testuser\", password: \"password123\", passwordConfirm: \"password123\" }) { ok token } }"
  }'
```

### Пример 2: Получить текущего пользователя (с cookie)

```bash
# Сначала залогиньтесь и получите токен, затем:
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: cnpf_auth=YOUR_TOKEN_HERE" \
  -d '{"query":"{ me { id email username } }"}'
```

### Пример 3: Получить список отчетов

```bash
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ reports(limit: 5) { id title author { username } } }"}'
```

## 📚 Полезные советы

1. **Используйте GraphQL Playground** - это самый удобный способ тестирования
2. **Проверяйте документацию схемы** - в Playground есть вкладка "Schema" с полным описанием
3. **Используйте переменные** - в Playground можно использовать вкладку "QUERY VARIABLES"
4. **Проверяйте ошибки** - GraphQL всегда возвращает структурированные ошибки в поле `errors`

## 🔍 Пример с переменными (в Playground)

**QUERY:**
```graphql
query GetReport($id: ID!) {
  report(id: $id) {
    id
    title
    text
    author {
      username
    }
  }
}
```

**QUERY VARIABLES:**
```json
{
  "id": "REPORT_ID_HERE"
}
```

## ⚠️ Частые ошибки

1. **"Не авторизован"** - нужно сначала выполнить `login` или `register`
2. **"Недостаточно прав"** - операция требует прав администратора
3. **"Отчет не найден"** - проверьте правильность ID
4. **"Email или имя пользователя уже используются"** - пользователь с такими данными уже существует

## 🎯 Быстрый старт для тестирования

1. Откройте `http://localhost:4000/` в браузере
2. Выполните регистрацию:
   ```graphql
   mutation {
     register(input: {
       email: "test@example.com"
       username: "testuser"
       password: "password123"
       passwordConfirm: "password123"
     }) {
       ok
       token
     }
   }
   ```
3. Проверьте текущего пользователя:
   ```graphql
   query {
     me {
       id
       email
       username
     }
   }
   ```
4. Создайте отчет:
   ```graphql
   mutation {
     createReport(input: {
       title: "Тестовый отчет"
       text: "Это тестовый отчет"
     }) {
       id
       title
     }
   }
   ```
5. Получите список отчетов:
   ```graphql
   query {
     reports(limit: 10) {
       id
       title
       author {
         username
       }
     }
   }
   ```

Готово! Теперь вы можете тестировать все запросы.
