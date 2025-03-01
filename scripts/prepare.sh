#!/bin/bash
set -e

go mod download

export PGPASSWORD="val1dat0r"

# Команды для PostgreSQL (пароль будет взят из переменной окружения)
psql -h localhost -p 5432 -U validator -d project-sem-1 <<-EOSQL
DROP TABLE IF EXISTS prices;
CREATE TABLE prices (
    product_id TEXT,
    creation_date DATE,
    product_name TEXT,
    category TEXT,
    price NUMERIC
);
GRANT ALL PRIVILEGES ON TABLE prices TO validator;
EOSQL

echo "Database setup completed successfully!"