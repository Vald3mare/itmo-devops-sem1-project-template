#!/bin/bash

# Установка PostgreSQL
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib

# Настройка PostgreSQL
sudo service postgresql start
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'postgres';"

# Создание пользователя, БД и таблицы
sudo -u postgres psql -c "CREATE USER validator WITH PASSWORD 'val1dat0r';"
sudo -u postgres psql -c "CREATE DATABASE \"project-sem-1\" OWNER validator;"
sudo -u postgres psql -d project-sem-1 -c "
    CREATE TABLE IF NOT EXISTS prices (
        product_id INTEGER,
        created_at DATE,
        product_name TEXT,
        category TEXT,
        price NUMERIC
    );"

# Настройка прав доступа
echo "host all all 127.0.0.1/32 md5" | sudo tee -a /etc/postgresql/*/main/pg_hba.conf
sudo service postgresql restart