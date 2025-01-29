#!/bin/bash

# Используем переменные окружения
export PGUSER=${POSTGRES_USER:-validator}
export PGPASSWORD=${POSTGRES_PASSWORD:-val1dat0r}
export PGDATABASE=${POSTGRES_DB:-project-sem-1}
export PGHOST=${POSTGRES_HOST:-localhost}
export PGPORT=${POSTGRES_PORT:-5432}

# Ждем готовности БД
for i in {1..10}; do
  pg_isready -h $PGHOST -p $PGPORT -U $PGUSER && break
  sleep 2
done

# Запуск приложения
go run cmd/main.go