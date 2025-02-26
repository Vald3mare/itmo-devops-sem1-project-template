package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// InitDB инициализирует пул соединений с PostgreSQL
func InitDB() (*pgxpool.Pool, error) {
	// Загрузка переменных окружения из файла .env
	err := godotenv.Load("database.env")
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
		return nil, err
	}

	// Получение переменных окружения
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Формирование строки подключения
	connString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbUser, dbPassword, dbName, dbHost, dbPort)

	// Создание пула соединений
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Ошибка создания пула соединений: %v", err)
		return nil, err
	}

	// Проверка подключения
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
		return nil, err
	}

	log.Println("Успешное подключение к базе данных!")
	return pool, nil
}
