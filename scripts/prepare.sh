#!/bin/bash

# Используем переменные окружения из GitHub Actions
PGHOST="localhost"
PGPORT="5432"
PGUSER="validator"
PGPASSWORD="val1dat0r"
PGDATABASE="project-sem-1"

# Ждем готовности PostgreSQL
until pg_isready -h $PGHOST -p $PGPORT -U $PGUSER; do
  echo "Waiting for PostgreSQL to start..."
  sleep 2
done

# Создаем таблицу (если не существует)
psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE << EOF
CREATE TABLE IF NOT EXISTS prices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    create_date DATE NOT NULL
);
EOF

# Проверка создания таблицы
psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -c "\dt+ prices"