# Проверка содержимого базы данных

## Способ 1: Через MongoDB Shell (mongosh)

### Подключиться к MongoDB в Docker

```bash
# Войти в MongoDB shell
docker compose exec mongo mongosh cnpf_feeder

# Или выполнить команду напрямую
docker compose exec mongo mongosh cnpf_feeder --eval "команда"
```

### Полезные команды

**Проверить количество пользователей:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.users.countDocuments()"
```

**Посмотреть всех пользователей:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.users.find({}, {email: 1, username: 1, isAdmin: 1, createdAt: 1}).toArray()"
```

**Проверить количество отчетов:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.reports.countDocuments()"
```

**Посмотреть отчеты:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.reports.find({}, {title: 1, authorId: 1, createdAt: 1}).limit(10).toArray()"
```

**Посмотреть все коллекции:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.getCollectionNames()"
```

**Посмотреть соревнования:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.competitions.find({}, {title: 1, startDate: 1, endDate: 1}).toArray()"
```

## Способ 2: Через GraphQL Playground

Откройте: `http://localhost:4000/`

### Запросы для проверки данных

**Проверить текущего пользователя:**
```graphql
query {
  me {
    id
    email
    username
    isAdmin
  }
}
```

**Посмотреть отчеты:**
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
    }
  }
}
```

**Посмотреть соревнования:**
```graphql
query {
  competitions {
    id
    title
    startDate
    endDate
  }
}
```

**Посмотреть пользователей (только для админов):**
```graphql
query {
  adminUsers {
    id
    email
    username
    isAdmin
  }
}
```

## Способ 3: Через curl

**Проверить пользователя:**
```bash
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ me { id email username } }"}'
```

**Посмотреть отчеты:**
```bash
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ reports(limit: 5) { id title author { username } } }"}'
```

## Способ 4: Интерактивный MongoDB Shell

```bash
# Войти в контейнер
docker compose exec mongo mongosh cnpf_feeder

# Затем выполнять команды:
> db.users.find().pretty()
> db.reports.find().pretty()
> db.competitions.find().pretty()

# Выйти
> exit
```

## Полезные MongoDB команды

**Очистить коллекцию (осторожно!):**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.users.deleteMany({})"
```

**Удалить конкретного пользователя (через скрипт):**
```bash
# Из проекта Frontend
./scripts/delete-user.sh test@example.com

# Или из проекта Backend
./scripts/delete-user.sh test@example.com

# Или напрямую через MongoDB
docker compose exec mongo mongosh cnpf_feeder --eval "db.users.deleteOne({email: 'test@example.com'})"
```

**Найти пользователя по email:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.users.findOne({email: 'test@example.com'})"
```

**Статистика базы данных:**
```bash
docker compose exec mongo mongosh cnpf_feeder --eval "db.stats()"
```
