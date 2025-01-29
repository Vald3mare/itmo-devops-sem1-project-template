#!/bin/bash

# Ждем готовности PostgreSQL
until pg_isready -h localhost -p 5432 -U validator; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

# Создание таблиц
psql -h localhost -U validator -d project-sem-1 << EOF
CREATE TABLE IF NOT EXISTS prices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    create_date DATE NOT NULL
);
EOF

# Права пользователя
psql -h localhost -U validator -d project-sem-1 -c "GRANT ALL PRIVILEGES ON TABLE prices TO validator;"