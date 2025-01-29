package db

import (
	"context"
	"fmt"
	//в практикуме рекомендовали использоать этот драй
	"github.com/jackc/pgx/v5"
)

// InitDB устанавливает подключение к базе данных
func InitDB() (*pgx.Conn, error) {
	connStr := "postgres://validator:val1dat0r@localhost:5432/project-sem-1?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка при проверке подключения к БД: %w", err)
	}

	return conn, nil
}
