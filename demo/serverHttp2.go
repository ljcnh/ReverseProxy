package main

import (
	"log"
	"net/http"
)

func ServeHTTP2(w http.ResponseWriter, r *http.Request) {
	log.Printf("Server2")
	w.Write([]byte(string("Server2\n")))
}

func main() {
	http.HandleFunc("/", ServeHTTP2)
	http.ListenAndServe("127.0.0.1:1332", nil)
}
