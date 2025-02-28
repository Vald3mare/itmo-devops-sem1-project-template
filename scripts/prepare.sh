#!/bin/bash

# Определение пути до файла database.env
ENV_FILE="database.env"

# Загрузка переменных окружения из файла database.env
if [ -f "$ENV_FILE" ]; then
  echo "Загрузка переменных окружения из файла $ENV_FILE..."
  source "$ENV_FILE"
else
  echo "Файл $ENV_FILE не найден."
  exit 1
fi

echo "Установка зависимостей..."
go mod download

# Настройка базы данных PostgreSQL
echo "Настройка базы данных PostgreSQL..."

# Создание базы данных, если она не существует
echo "Проверка установки PostgreSQL..."
if pg_isready -q -h "$PGHOST" -p "$PGPORT" -U "$PGUSER"; then
    echo "PostgreSQL уже установлен."
else
    echo "PostgreSQL не установлен. Запуск установки..."
    sudo apt-get update
    sudo apt-get install -y golang-go postgresql postgresql-contrib unzip curl
    sudo service postgresql start
fi

# Проверка наличия базы данных project-sem-1
echo "Проверка наличия бд 'project-sem-1'..."
DB_EXISTS=$(PGPASSWORD="$PGPASSWORD" psql -U "$PGUSER" -h "$PGHOST" -p "$PGPORT" -d "$DBNAME" -tAc "SELECT 1 FROM pg_database WHERE datname='project-sem-1'")

if [ "$DB_EXISTS" == "1" ]; then
    echo "База данных 'project-sem-1' уже существует. Пропускаем процесс инициализации..."
else
    echo "Инициализируем базу данных 'project-sem-1'..."
    psql -U "$PGUSER" -h "$PGHOST" -p "$PGPORT" -d "$DBNAME" <<EOF
    CREATE DATABASE "project-sem-1";
    CREATE USER validator WITH PASSWORD 'val1dat0r';
    GRANT ALL PRIVILEGES ON DATABASE "project-sem-1" TO validator;
EOF
fi

# Проверка наличия таблицы prices
echo "Checking if table 'prices' exists..."
TABLE_EXISTS=$(PGPASSWORD="$PGPASSWORD" psql -U "$PGUSER" -h "$PGHOST" -p "$PGPORT" -d "$DBNAME" -tAc "SELECT to_regclass('public.prices')")

if [ "$TABLE_EXISTS" == "public.prices" ]; then
    echo "Таблица 'prices' уже существует. Пропускаем инициализацию..."
else
    echo "Инициализация таблицы 'prices'..."
    PGPASSWORD="$PGPASSWORD" psql -U "$PGUSER" -h "$PGHOST" -p "$PGPORT" -d "$DBNAME" <<EOF
    ALTER SCHEMA public OWNER TO validator;
    GRANT ALL ON SCHEMA public TO validator;

    CREATE TABLE prices (
        id SERIAL PRIMARY KEY,
        product_name TEXT NOT NULL,
        category TEXT NOT NULL,
        price NUMERIC NOT NULL,
        creation_date timestamp NOT NULL
    );

    GRANT ALL PRIVILEGES ON TABLE prices TO validator;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO validator;
EOF
fi

echo "Подготовка окружения завершена."