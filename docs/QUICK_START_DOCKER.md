# 🚀 Быстрый запуск Backend через Docker

## Шаг 1: Очистка Docker

```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# Остановить все контейнеры
docker compose down

# Удалить старые образы
docker rmi cnpf-feeder-backend 2>/dev/null || true

# Очистить кэш сборки
docker builder prune -f
```

## Шаг 2: Подготовка проекта

```bash
# Обновить зависимости
go mod tidy

# Сгенерировать GraphQL код локально (чтобы убедиться что все работает)
go generate ./graph

# Проверить компиляцию
go build -o /tmp/test ./cmd/graph && echo "✅ Локальная сборка успешна"
```

## Шаг 3: Создать .env файл

```bash
# Создать .env с секретным ключом
cat > .env << 'EOF'
AUTH_SECRET=$(openssl rand -hex 32)
MONGODB_URI=mongodb://mongo:27017/cnpf_feeder
PORT=4000
GIN_MODE=release
CORS_ORIGIN=http://localhost:3000
EOF

# Или вручную отредактируйте .env и укажите AUTH_SECRET
```

**Важно:** `AUTH_SECRET` должен быть минимум 32 символа!

## Шаг 4: Пересобрать Docker образ

```bash
# Пересобрать без кэша
docker compose build --no-cache backend
```

## Шаг 5: Запустить Backend

```bash
# Запустить MongoDB и Backend
docker compose up mongo backend

# Или в фоновом режиме
docker compose up -d mongo backend
```

## Проверка работы

```bash
# Проверить логи
docker compose logs backend

# Проверить health endpoint
curl http://localhost:4000/health

# Открыть GraphQL Playground (если GIN_MODE=debug)
open http://localhost:4000/
```

## Остановка

```bash
docker compose down
```

---

## 🔧 Если что-то пошло не так

### Проблема: "missing AUTH_SECRET"
```bash
# Проверьте что .env файл существует и содержит AUTH_SECRET
cat .env | grep AUTH_SECRET

# Или передайте через переменную окружения
export AUTH_SECRET=your-secret-key-here-min-32-chars
docker compose up mongo backend
```

### Проблема: "Failed to connect to MongoDB"
```bash
# Убедитесь что MongoDB запущен
docker compose ps

# Проверьте логи MongoDB
docker compose logs mongo
```

### Проблема: Ошибка сборки Docker
```bash
# Полная очистка и пересборка
docker compose down -v
docker system prune -af
docker compose build --no-cache backend
```

---

## 📝 Полная команда одной строкой

```bash
cd /Users/daniel/projects/cnpf-feeder-backend && \
go mod tidy && \
go generate ./graph && \
docker compose down && \
docker compose build --no-cache backend && \
export AUTH_SECRET=$(openssl rand -hex 32) && \
docker compose up mongo backend
```
