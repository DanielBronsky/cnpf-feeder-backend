#!/bin/bash

# Скрипт для запуска Backend через Docker

# Не останавливаться на ошибках (кроме критических)
set +e

echo "🧹 Шаг 1: Очистка Docker..."
docker compose down 2>/dev/null || true
docker rmi cnpf-feeder-backend 2>/dev/null || true
docker builder prune -f

echo "📦 Шаг 2: Подготовка проекта..."
go mod tidy
go mod download

echo "🔧 Шаг 3: Генерация GraphQL кода (обязательно перед Docker сборкой)..."
if cd graph && go run github.com/99designs/gqlgen@latest generate && cd ..; then
    echo "✅ GraphQL код сгенерирован локально"
else
    echo "❌ Ошибка генерации GraphQL кода!"
    echo "Попробуйте вручную: cd graph && go run github.com/99designs/gqlgen@latest generate"
    exit 1
fi

echo "✅ Шаг 4: Проверка локальной сборки..."
if go build -o /tmp/test-build ./cmd/graph 2>/dev/null; then
    rm /tmp/test-build
    echo "✅ Локальная сборка успешна"
else
    echo "⚠️  Локальная сборка пропущена (будет выполнена в Docker)"
fi

echo "🐳 Шаг 5: Создание .env файла..."
if [ ! -f .env ]; then
    AUTH_SECRET=$(openssl rand -hex 32)
    cat > .env << EOF
AUTH_SECRET=${AUTH_SECRET}
MONGODB_URI=mongodb://mongo:27017/cnpf_feeder
PORT=4000
GIN_MODE=release
CORS_ORIGIN=http://localhost:3000
EOF
    echo "✅ .env файл создан с AUTH_SECRET"
else
    echo "ℹ️  .env файл уже существует"
fi

echo "🔨 Шаг 6: Сборка Docker образа..."
if ! docker compose build --no-cache backend; then
    echo "❌ Ошибка сборки Docker образа!"
    exit 1
fi

echo "🚀 Шаг 7: Запуск Backend..."
docker compose up mongo backend
