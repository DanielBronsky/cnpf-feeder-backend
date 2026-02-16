#!/bin/bash
# Скрипт для удаления пользователя из MongoDB по email
# Использование: ./scripts/delete-user.sh <email>

if [ -z "$1" ]; then
  echo "Использование: $0 <email>"
  echo "Пример: $0 danielbronsky4@gmail.com"
  exit 1
fi

EMAIL="$1"

# Находим запущенный MongoDB контейнер
MONGO_CONTAINER=$(docker ps --filter "name=mongo" --format "{{.Names}}" | head -n 1)

if [ -z "$MONGO_CONTAINER" ]; then
  echo "Ошибка: MongoDB контейнер не найден"
  echo "Проверьте запущенные контейнеры: docker ps | grep mongo"
  echo "Или запустите: docker compose -f docker-compose.dev.yml up mongo -d"
  exit 1
fi

echo "Найден MongoDB контейнер: $MONGO_CONTAINER"
echo "Поиск пользователя с email: $EMAIL"

# Удаляем пользователя через найденный контейнер
docker exec "$MONGO_CONTAINER" mongosh cnpf_feeder --quiet --eval "
const user = db.users.findOne({ email: '$EMAIL' }, { email: 1, username: 1, _id: 1 });
if (user) {
  print('Найден пользователь:');
  print(JSON.stringify(user, null, 2));
  const result = db.users.deleteOne({ email: '$EMAIL' });
  print('Результат удаления:');
  print(JSON.stringify(result, null, 2));
  if (result.deletedCount === 1) {
    print('✅ Пользователь успешно удален');
  } else {
    print('⚠️ Пользователь не был удален');
  }
} else {
  print('❌ Пользователь с email \"$EMAIL\" не найден');
}
"
