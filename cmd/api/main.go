package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"project-sem/internal/db"
	"project-sem/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Инициализация пула соединений с базой данных
	pool, err := db.InitDB()
	if err != nil {
		return err
	}
	defer pool.Close()

	// Инициализация роутера
	r := mux.NewRouter()

	// Передача пула соединений в хендлеры
	r.HandleFunc("/api/v0/prices", handlers.HandlerPostPrices(pool)).Methods("POST")
	r.HandleFunc("/api/v0/prices", handlers.HandlerGetPrices(pool)).Methods("GET")

	// Запуск сервера
	err = http.ListenAndServe(`:80`, r)
	if err != nil {
		return err
	}
	return nil
}
