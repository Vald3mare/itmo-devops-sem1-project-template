#!/bin/bash

# Переменные окружения
export PGUSER="${POSTGRES_USER:-validator}"
export PGPASSWORD="${POSTGRES_PASSWORD:-val1dat0r}"
export PGDATABASE="${POSTGRES_DB:-project-sem-1}"
export PGHOST="${POSTGRES_HOST:-localhost}"
export PGPORT="${POSTGRES_PORT:-5432}"

# Сборка приложения
echo "🔨 Сборка приложения..."
go build -o main cmd/main.go

# Ожидание БД
for i in {1..15}; do
  echo "Проверка подключения к БД ($i/15)..."
  if pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER"; then
    echo "База данных готова!"
    break
  fi
  sleep 2
done

# Запуск приложения
echo "Запуск приложения..."
exec ./main