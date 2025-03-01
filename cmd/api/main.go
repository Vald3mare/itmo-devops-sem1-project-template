package main

import (
	"net/http"
	"project-sem/internal/handlers"
	"project-sem/internal/myDB"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := myDB.InitDB(); err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v0/prices", handlers.HandlerPostPrices()).Methods("POST")
	r.HandleFunc("/api/v0/prices", handlers.HandlerGetPrices()).Methods("GET")

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		return err
	}
	return nil
}
