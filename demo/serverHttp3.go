package main

import (
	"log"
	"net/http"
)

func ServeHTTP3(w http.ResponseWriter, r *http.Request) {
	log.Printf("default")
	w.Write([]byte(string("default\n")))
}

func main() {
	http.HandleFunc("/", ServeHTTP3)
	http.ListenAndServe("127.0.0.1:1333", nil)
}
