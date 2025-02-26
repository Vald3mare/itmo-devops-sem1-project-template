package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func InitDB() (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return pool, nil
}
