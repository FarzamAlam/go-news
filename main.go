package main

import (
	"log"
	"net/http"
	"os"

	"github.com/farzamalam/go-news/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()
	// Serving static files
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search", handlers.SearchHandler)
	mux.HandleFunc("/", handlers.IndexHandler)
	log.Println("Starting server on : ", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
