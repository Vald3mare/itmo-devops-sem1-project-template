#!/bin/bash

# Установка PostgreSQL
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib

# Запуск PostgreSQL
sudo service postgresql start

# Создание пользователя и БД
sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"

# Создание таблицы
sudo -u postgres psql -d project-sem-1 -c "CREATE TABLE IF NOT EXISTS prices (
    id TEXT,
    created_at DATE,
    product_name TEXT,
    category TEXT,
    price DECIMAL
);"

# Права пользователя
sudo -u postgres psql -d project-sem-1 -c "GRANT ALL PRIVILEGES ON TABLE prices TO validator;"