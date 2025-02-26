#!/bin/bash

PGHOST="${POSTGRES_HOST:-localhost}"
PGPORT="${POSTGRES_PORT:-5432}"
PGUSER="${POSTGRES_USER:-validator}"
PGPASSWORD="${POSTGRES_PASSWORD:-val1dat0r}"
PGDATABASE="${POSTGRES_DB:-project-sem-1}"

# Проверка доступности PostgreSQL
for i in {1..15}; do
  echo "Checking PostgreSQL ($i/15)..."
  if pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER"; then
    echo "PostgreSQL ready!"
    break
  fi
  sleep 2
done

# Создание таблицы prices
psql "postgresql://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE" << EOF
CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL,
    product_name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    created_at DATE NOT NULL
);
EOF

# Проверка создания таблицы
psql "postgresql://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE" -c "\dt+ prices"
if [ $? -ne 0 ]; then
  echo "Ошибка при создании таблицы prices"
  exit 1
fi

echo "Таблица prices успешно создана!"