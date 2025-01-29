#!/bin/bash

# Установка PostgreSQL
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib

# Настройка пользователя и БД
sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"
sudo -u postgres psql -d project-sem-1 -c "GRANT ALL PRIVILEGES ON DATABASE \"project-sem-1\" TO validator;"

# Создание таблицы prices
sudo -u postgres psql -d project-sem-1 << EOF
CREATE TABLE IF NOT EXISTS prices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    create_date DATE NOT NULL
);
EOF

# Права для пользователя
sudo -u postgres psql -d project-sem-1 -c "GRANT ALL PRIVILEGES ON TABLE prices TO validator;"