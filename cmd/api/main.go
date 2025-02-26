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
	pool, err := db.InitDB()
	if err != nil {
		return err
	}
	defer pool.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/v0/prices", handlers.HandlerPostPrices(pool)).Methods("POST")
	r.HandleFunc("/api/v0/prices", handlers.HandlerGetPrices(pool)).Methods("GET")

	err = http.ListenAndServe(`:8080`, r)
	if err != nil {
		return err
	}
	return nil
}
