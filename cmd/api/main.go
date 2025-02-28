package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"project-sem/internal/handlers"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "validator"
	dbPassword = "val1dat0r"
	dbName     = "project-sem-1"
)

var db *sql.DB

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var err error
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := InitDB(); err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v0/prices", handlers.HandlerPostPrices()).Methods("POST")
	r.HandleFunc("/api/v0/prices", handlers.HandlerGetPrices()).Methods("GET")

	err = http.ListenAndServe(`:8080`, r)
	if err != nil {
		return err
	}
	return nil
}

func InitDB() error {
	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return err
	}

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS prices (
        id SERIAL PRIMARY KEY,
        product_name TEXT NOT NULL,
        category TEXT NOT NULL,
        price NUMERIC(10,2) NOT NULL,
        creation_date timestamp NOT NULL
    )`)
	return err
}
