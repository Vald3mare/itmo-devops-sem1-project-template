#!/bin/bash

go mod tidy

# Подготовка базы данных
psql -U validator -d project-sem-1 -c "CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    product_id INT,
    created_at DATE,
    product_name TEXT,
    category TEXT,
    price DECIMAL(10, 2)
);"