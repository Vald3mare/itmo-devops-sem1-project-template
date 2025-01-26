package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)
var conn *pgxpool.Conn

func ConnectToDB() {
	connStr := "postgres://validator:val1dat0r@localhost:5432/project-sem-1?sslmode=disable"
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	conn, err = dbpool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a connection: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("DB connected successfully!")
}