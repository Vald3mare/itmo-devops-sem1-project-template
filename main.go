package main

import (
	"net/http"
)

func helloFunc(w http.ResponseWriter, req *http.Request){
	w.Write([]byte("Hello, world!"))
}

func main() {
	mux := http.NewServeMux()
	http.HandleFunc("/", helloFunc)
	http.ListenAndServe("localhost:80", mux)
}
