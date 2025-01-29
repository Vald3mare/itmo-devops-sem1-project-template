#!/bin/bash

# Переменные окружения (можно переопределить в CI)
PGHOST="${POSTGRES_HOST:-localhost}"
PGPORT="${POSTGRES_PORT:-5432}"
PGUSER="${POSTGRES_USER:-validator}"
PGPASSWORD="${POSTGRES_PASSWORD:-val1dat0r}"
PGDATABASE="${POSTGRES_DB:-project-sem-1}"

# Ожидание готовности PostgreSQL
for i in {1..15}; do
  echo "Checking PostgreSQL ($i/15)..."
  if pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER"; then
    echo "✅ PostgreSQL ready!"
    break
  fi
  sleep 2
done

# Создание таблицы
psql "postgresql://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE" << EOF
CREATE TABLE IF NOT EXISTS prices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    create_date DATE NOT NULL
);
EOF

# Проверка
psql "postgresql://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE" -c "\dt+ prices"