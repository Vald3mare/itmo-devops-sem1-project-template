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
	err := db.InitDB()
	if err != nil {
		return err
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/v0/prices", handlers.HandlerPostPrices()).Methods("POST")
	r.HandleFunc("/api/v0/prices", handlers.HandlerGetPrices()).Methods("GET")
	err = http.ListenAndServe(`:80`, r)
	if err != nil {
		return err
	}
	return nil
}
