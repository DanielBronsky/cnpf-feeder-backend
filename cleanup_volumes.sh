#!/bin/bash

# Скрипт для безопасной очистки неиспользуемых volumes

set -e

echo "🔍 Проверка volumes..."

# Получаем список всех volumes
ALL_VOLUMES=$(docker volume ls -q)

# Получаем volumes, которые используются активными контейнерами
USED_VOLUMES=$(docker ps --format "{{.Names}}" | xargs -I {} docker inspect {} --format '{{range .Mounts}}{{if .Name}}{{.Name}} {{end}}{{end}}' 2>/dev/null | tr ' ' '\n' | grep -v '^$' | sort -u)

echo ""
echo "📦 Используемые volumes (НЕ будут удалены):"
for vol in $USED_VOLUMES; do
    if [ ! -z "$vol" ]; then
        SIZE=$(docker system df -v | grep "$vol" | awk '{print $3}' || echo "unknown")
        echo "   ✅ $vol ($SIZE)"
    fi
done

echo ""
echo "🗑️  Неиспользуемые volumes (будут удалены):"
UNUSED_COUNT=0
for vol in $ALL_VOLUMES; do
    if echo "$USED_VOLUMES" | grep -q "^${vol}$"; then
        continue
    fi
    SIZE=$(docker system df -v 2>/dev/null | grep "$vol" | awk '{print $3}' || echo "unknown")
    echo "   ❌ $vol ($SIZE)"
    UNUSED_COUNT=$((UNUSED_COUNT + 1))
done

if [ $UNUSED_COUNT -eq 0 ]; then
    echo "   ✅ Неиспользуемых volumes нет"
    exit 0
fi

echo ""
read -p "Удалить $UNUSED_COUNT неиспользуемых volumes? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Отменено"
    exit 0
fi

echo ""
echo "🗑️  Удаление неиспользуемых volumes..."
for vol in $ALL_VOLUMES; do
    if echo "$USED_VOLUMES" | grep -q "^${vol}$"; then
        continue
    fi
    echo "   Удаление: $vol"
    docker volume rm "$vol" 2>/dev/null || echo "     ⚠️  Не удалось удалить (возможно используется)"
done

echo ""
echo "🧹 Очистка неиспользуемых volumes через prune..."
docker volume prune -f

echo ""
echo "📊 Итоговая статистика volumes:"
docker volume ls

echo ""
echo "💾 Использование диска:"
docker system df

echo ""
echo "✅ Очистка volumes завершена!"
