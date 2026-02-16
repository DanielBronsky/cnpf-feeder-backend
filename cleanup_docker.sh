#!/bin/bash

# Скрипт для очистки Docker от лишних контейнеров и образов

set -e

echo "🧹 Очистка Docker..."

echo ""
echo "📦 Шаг 1: Удаление остановленных контейнеров..."
STOPPED=$(docker ps -a --filter "status=exited" -q)
if [ -z "$STOPPED" ]; then
    echo "   ✅ Остановленных контейнеров нет"
else
    echo "   Найдено остановленных контейнеров: $(echo $STOPPED | wc -w)"
    docker rm $STOPPED
    echo "   ✅ Остановленные контейнеры удалены"
fi

echo ""
echo "🗑️  Шаг 2: Удаление dangling images (образы без тегов)..."
DANGLING=$(docker images --filter "dangling=true" -q)
if [ -z "$DANGLING" ]; then
    echo "   ✅ Dangling images нет"
else
    echo "   Найдено dangling images: $(echo $DANGLING | wc -w)"
    docker rmi $DANGLING 2>/dev/null || true
    echo "   ✅ Dangling images удалены"
fi

echo ""
echo "🗑️  Шаг 3: Удаление старых образов backend/frontend..."
# Удаляем старые образы с префиксом cnpffeedermd (старое название проекта)
OLD_BACKEND=$(docker images cnpffeedermd-backend -q)
OLD_FRONTEND=$(docker images cnpffeedermd-frontend -q)

if [ ! -z "$OLD_BACKEND" ]; then
    echo "   Удаление cnpffeedermd-backend..."
    docker rmi $OLD_BACKEND 2>/dev/null || true
    echo "   ✅ Старый backend образ удален"
fi

if [ ! -z "$OLD_FRONTEND" ]; then
    echo "   Удаление cnpffeedermd-frontend..."
    docker rmi $OLD_FRONTEND 2>/dev/null || true
    echo "   ✅ Старый frontend образ удален"
fi

echo ""
echo "🧹 Шаг 4: Очистка неиспользуемых ресурсов..."
docker system prune -f

echo ""
echo "📊 Итоговая статистика:"
echo "   Контейнеры:"
docker ps -a --format "   - {{.Names}} ({{.Status}})"
echo ""
echo "   Образы:"
docker images --format "   - {{.Repository}}:{{.Tag}} ({{.Size}})"
echo ""
echo "💾 Использование диска:"
docker system df

echo ""
echo "✅ Очистка завершена!"
