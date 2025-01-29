package db

import (
	"context"
	"fmt"
	"os"
	//в практикуме рекомендовали использоать этот драй
	"github.com/jackc/pgx/v5"
)

// InitDB устанавливает подключение к базе данных
func InitDB() (*pgx.Conn, error) {
	connStr := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_HOST"),
        os.Getenv("POSTGRES_PORT"),
        os.Getenv("POSTGRES_DB"),
    )
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка при проверке подключения к БД: %w", err)
	}

	return conn, nil
}
