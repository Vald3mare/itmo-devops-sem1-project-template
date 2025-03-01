#!/bin/bash
set -e

go mod download

psql -U validator -d project-sem-1 -h localhost -p 5432 <<-EOSQL
DROP TABLE IF EXISTS prices;
CREATE TABLE IF NOT EXISTS prices (
    product_id TEXT,
    creation_date DATE,
    product_name TEXT,
    category TEXT,
    price NUMERIC
);

GRANT ALL PRIVILEGES ON TABLE prices TO validator;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO validator;
EOSQL