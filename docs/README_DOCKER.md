# 🐳 Запуск Backend через Docker

## Быстрый старт (автоматический)

Просто запустите скрипт:
```bash
./START_BACKEND.sh
```

Скрипт автоматически:
1. Очистит старые Docker образы
2. Обновит зависимости
3. Сгенерирует GraphQL код
4. Создаст .env файл
5. Соберет Docker образ
6. Запустит Backend

---

## Ручной запуск

### 1. Очистка Docker

```bash
docker compose down
docker rmi cnpf-feeder-backend 2>/dev/null || true
docker builder prune -f
```

### 2. Подготовка проекта

```bash
go mod tidy
go mod download
```

### 3. Генерация GraphQL кода

```bash
cd graph
go run github.com/99designs/gqlgen generate
cd ..
```

### 4. Создание .env файла

```bash
# Автоматически сгенерировать секретный ключ
AUTH_SECRET=$(openssl rand -hex 32)

cat > .env << EOF
AUTH_SECRET=${AUTH_SECRET}
MONGODB_URI=mongodb://mongo:27017/cnpf_feeder
PORT=4000
GIN_MODE=release
CORS_ORIGIN=http://localhost:3000
EOF
```

### 5. Сборка Docker образа

```bash
docker compose build --no-cache backend
```

### 6. Запуск

```bash
# Запустить MongoDB и Backend
docker compose up mongo backend

# Или в фоновом режиме
docker compose up -d mongo backend
```

---

## Проверка работы

```bash
# Логи Backend
docker compose logs -f backend

# Health check
curl http://localhost:4000/health

# GraphQL Playground (если GIN_MODE=debug)
open http://localhost:4000/
```

---

## Остановка

```bash
docker compose down
```

---

## Решение проблем

### Ошибка: "missing AUTH_SECRET"
```bash
# Проверить .env файл
cat .env | grep AUTH_SECRET

# Или передать через переменную окружения
export AUTH_SECRET=$(openssl rand -hex 32)
docker compose up mongo backend
```

### Ошибка сборки Docker
```bash
# Полная очистка
docker compose down -v
docker system prune -af
docker builder prune -af

# Пересборка
docker compose build --no-cache backend
```

### MongoDB не подключается
```bash
# Проверить что MongoDB запущен
docker compose ps

# Проверить логи MongoDB
docker compose logs mongo

# Проверить сеть
docker network ls | grep cnpf
```
