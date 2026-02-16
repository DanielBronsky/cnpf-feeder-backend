# Установка MongoDB для разработки

## Проблема с Docker

Если MongoDB в Docker падает с ошибкой "No space left on device", это означает, что Docker использует слишком много места или имеет ограничения.

## Решение: MongoDB локально через Homebrew

### Установка

```bash
# Добавить tap для MongoDB
brew tap mongodb/brew

# Установить MongoDB Community Edition
brew install mongodb-community

# Запустить MongoDB как сервис
brew services start mongodb-community

# Или запустить вручную (без автозапуска)
mongod --config /opt/homebrew/etc/mongod.conf
```

### Проверка

```bash
# Проверить статус
brew services list | grep mongo

# Или проверить подключение
mongosh
```

### Остановка

```bash
# Остановить сервис
brew services stop mongodb-community
```

### Настройка Backend

После установки MongoDB локально, ваш `.env` файл уже настроен правильно:

```env
MONGODB_URI="mongodb://localhost:27017/cnpf_feeder"
```

Просто запустите Backend:

```bash
go run ./cmd/server
```

## Альтернатива: MongoDB Atlas (облако)

Если не хотите устанавливать MongoDB локально, можно использовать бесплатный кластер MongoDB Atlas:

1. Зарегистрируйтесь на https://www.mongodb.com/cloud/atlas
2. Создайте бесплатный кластер (M0)
3. Получите connection string
4. Обновите `.env`:

```env
MONGODB_URI="mongodb+srv://username:password@cluster.mongodb.net/cnpf_feeder?retryWrites=true&w=majority"
```

## Очистка Docker (если нужно)

Если хотите освободить место в Docker:

```bash
# Удалить неиспользуемые образы, контейнеры, volumes
docker system prune -a --volumes

# Удалить только build cache (освободит ~21GB)
docker builder prune -a -f
```

⚠️ **Внимание**: Это удалит все неиспользуемые ресурсы Docker!
