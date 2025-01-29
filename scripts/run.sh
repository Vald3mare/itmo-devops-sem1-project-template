#!/bin/bash

# Ждем готовности БД
while ! pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER; do
  echo "Waiting for database..."
  sleep 2
done

# Запуск приложения
go run cmd/main.go