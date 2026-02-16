# Команды для работы с Docker

## 🧹 1. Очистка Docker

### Удалить старые образы и контейнеры
```bash
# Остановить и удалить все контейнеры проекта
docker compose down

# Удалить старые образы
docker rmi cnpf-feeder-backend 2>/dev/null || true
docker rmi $(docker images | grep cnpf-feeder-backend | awk '{print $3}') 2>/dev/null || true

# Очистить неиспользуемые образы и кэш
docker system prune -f

# Очистить build cache (опционально, если проблемы с кэшем)
docker builder prune -af
```

## 🔨 2. Пересборка проекта локально

### Обновить зависимости и сгенерировать GraphQL код
```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# Обновить зависимости
go mod tidy
go mod download

# Сгенерировать GraphQL код
go generate ./graph

# Проверить что все компилируется
go build -o /tmp/test-build ./cmd/graph
```

## 🐳 3. Пересборка Docker образа

### Вариант А: Через docker-compose (рекомендуется)
```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# Пересобрать образ без кэша
docker compose build --no-cache backend

# Или пересобрать все сервисы
docker compose build --no-cache
```

### Вариант Б: Напрямую через docker build
```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# Пересобрать образ
docker build --no-cache -t cnpf-feeder-backend .
```

## 🚀 4. Запуск Backend через Docker

### Вариант А: Через docker-compose (с MongoDB)
```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# Создать .env файл если его нет
cat > .env << EOF
AUTH_SECRET=your-super-secret-key-minimum-32-characters-long-change-in-production
MONGODB_URI=mongodb://mongo:27017/cnpf_feeder
PORT=4000
GIN_MODE=release
CORS_ORIGIN=http://localhost:3000
EOF

# Запустить только backend и mongo
docker compose up mongo backend

# Или в фоне
docker compose up -d mongo backend
```

### Вариант Б: Только Backend (если MongoDB уже запущен)
```bash
# Запустить контейнер с переменными окружения
docker run -d \
  --name cnpf-feeder-backend \
  --network cnpf-network \
  -p 4000:4000 \
  -e AUTH_SECRET=your-super-secret-key-minimum-32-characters-long \
  -e MONGODB_URI=mongodb://mongo:27017/cnpf_feeder \
  -e PORT=4000 \
  -e GIN_MODE=release \
  -e CORS_ORIGIN=http://localhost:3000 \
  cnpf-feeder-backend
```

### Вариант В: Полный стек (Frontend + Backend + MongoDB)
```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# Убедитесь что .env файл создан с AUTH_SECRET
export AUTH_SECRET=your-super-secret-key-minimum-32-characters-long

# Запустить все сервисы
docker compose up --build
```

## 📋 5. Проверка работы

### Проверить логи
```bash
# Логи backend
docker compose logs backend

# Логи в реальном времени
docker compose logs -f backend

# Или если запущен напрямую
docker logs -f cnpf-feeder-backend
```

### Проверить что сервер работает
```bash
# Health check
curl http://localhost:4000/health

# GraphQL Playground (если GIN_MODE=debug)
open http://localhost:4000/
```

## 🛑 Остановка

```bash
# Остановить все сервисы
docker compose down

# Остановить и удалить volumes (данные MongoDB будут удалены!)
docker compose down -v

# Остановить конкретный контейнер
docker stop cnpf-feeder-backend
docker rm cnpf-feeder-backend
```

## 🔍 Отладка

### Зайти внутрь контейнера
```bash
docker compose exec backend sh

# Или если запущен напрямую
docker exec -it cnpf-feeder-backend sh
```

### Проверить переменные окружения в контейнере
```bash
docker compose exec backend env | grep -E "(AUTH_SECRET|MONGODB_URI|PORT)"
```

## ⚡ Быстрый старт (все команды подряд)

```bash
cd /Users/daniel/projects/cnpf-feeder-backend

# 1. Очистка
docker compose down
docker system prune -f

# 2. Локальная подготовка
go mod tidy
go generate ./graph

# 3. Создать .env
echo "AUTH_SECRET=$(openssl rand -hex 32)" > .env
echo "MONGODB_URI=mongodb://mongo:27017/cnpf_feeder" >> .env
echo "PORT=4000" >> .env
echo "GIN_MODE=release" >> .env
echo "CORS_ORIGIN=http://localhost:3000" >> .env

# 4. Пересобрать и запустить
docker compose build --no-cache backend
docker compose up mongo backend
```
