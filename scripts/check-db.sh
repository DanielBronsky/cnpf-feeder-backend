#!/bin/bash
# Скрипт для проверки содержимого базы данных

echo "📊 Статистика базы данных cnpf_feeder"
echo "======================================"
echo ""

echo "👥 Пользователи:"
docker compose exec -T mongo mongosh cnpf_feeder --quiet --eval "db.users.countDocuments()" 2>/dev/null | grep -v "Current Mongosh"
echo ""

echo "📝 Отчеты:"
docker compose exec -T mongo mongosh cnpf_feeder --quiet --eval "db.reports.countDocuments()" 2>/dev/null | grep -v "Current Mongosh"
echo ""

echo "🏆 Соревнования:"
docker compose exec -T mongo mongosh cnpf_feeder --quiet --eval "db.competitions.countDocuments()" 2>/dev/null | grep -v "Current Mongosh"
echo ""

echo "📋 Коллекции в базе:"
docker compose exec -T mongo mongosh cnpf_feeder --quiet --eval "db.getCollectionNames()" 2>/dev/null | grep -v "Current Mongosh"
echo ""

echo "👤 Последние 5 пользователей:"
docker compose exec -T mongo mongosh cnpf_feeder --quiet --eval "db.users.find({}, {email: 1, username: 1, isAdmin: 1, createdAt: 1}).sort({createdAt: -1}).limit(5).toArray()" 2>/dev/null | grep -v "Current Mongosh"
echo ""

echo "📄 Последние 5 отчетов:"
docker compose exec -T mongo mongosh cnpf_feeder --quiet --eval "db.reports.find({}, {title: 1, authorId: 1, createdAt: 1}).sort({createdAt: -1}).limit(5).toArray()" 2>/dev/null | grep -v "Current Mongosh"
