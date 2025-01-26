package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"project_sem/internal/db"
	"project_sem/internal/handlers"
)


func main() {
	db.ConnectToDB()

	router := mux.NewRouter()
	router.HandleFunc("/api/v0/prices", handlers.PostPrices).Methods("POST")
	router.HandleFunc("/api/v0/prices", handlers.GetPrices).Methods("GET")
	http.HandleFunc("/api/v0/prices", nil)
	http.ListenAndServe(":80", nil)
}
