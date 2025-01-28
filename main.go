package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"project-sem/internal/db"
	"project-sem/internal/handlers"
)

func main() {
	conn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Не удается подключиться к БД: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/api/v0/prices", handlers.HandlerGetPrices(conn)).Methods("GET")
	router.HandleFunc("/api/v0/prices", handlers.HandlerPostPrices(conn)).Methods("POST")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
