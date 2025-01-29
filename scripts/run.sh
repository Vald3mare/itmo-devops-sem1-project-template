#!/bin/bash

# Экспорт переменных окружения для PostgreSQL
export PGUSER=validator
export PGPASSWORD=val1dat0r
export PGDATABASE=project-sem-1
export PGHOST=localhost
export PGPORT=5432

# Запуск приложения
go run cmd/main.go