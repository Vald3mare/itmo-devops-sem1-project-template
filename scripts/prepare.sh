#!/bin/bash
set -e

export PGPASSWORD="val1dat0r"

# Установка зависимостей
go mod download

# Пересоздание таблицы
psql -h localhost -p 5432 -U validator -d project-sem-1 <<-EOSQL
	DROP TABLE IF EXISTS prices;
	CREATE TABLE prices (
		id TEXT,
		name TEXT,
		category TEXT,
		price NUMERIC,
		create_date DATE
	);
	GRANT ALL PRIVILEGES ON TABLE prices TO validator;
EOSQL

echo "Database prepared successfully!"