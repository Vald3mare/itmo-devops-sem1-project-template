package main

import (
	"net/http"
	"github.com/gorilla/mux"
	
)


func main() {
	connectToDB()
	http.HandleFunc("/api/v0/prices", apiHandler)
	http.ListenAndServe(":80", nil)
}
