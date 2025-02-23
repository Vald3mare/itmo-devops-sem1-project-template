#!/bin/bash

# Абсолютный путь к database.env
ENV_FILE="/database.env"

# Загружаем переменные окружения из database.env файла
if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
else
    echo "Файл database.env не найден по пути: $ENV_FILE"
    exit 1
fi

# Проверяем, что все необходимые переменные заданы
if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo "Одна или несколько переменных окружения не заданы в database.env файле!"
    exit 1
fi

# Ожидание готовности PostgreSQL
for i in {1..15}; do
  echo "Checking PostgreSQL ($i/15)..."
  if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; then
    echo "PostgreSQL ready!"
    break
  fi
  sleep 2
done

# Создание таблицы
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" << EOF
CREATE TABLE IF NOT EXISTS prices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    create_date DATE NOT NULL
);
EOF

# Проверка
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -c "\dt+ prices"g