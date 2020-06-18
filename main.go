package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/farzamalam/go-news/handlers"
)

// TO DO:
// 0. Implement a running app.		--> Done.
// 1. Get top stories by default for user country or india.
// 2. Update the Header UI.		--> Done.
// 3. Refractor newsHandler.	--> Done.
// 4. Implement Concurrency.
// 5. Implement Caching.
// 6. Deploy on heruko.			--> Done.
// 7. Make a proper README.
// 8. Post it on reddit.

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	r := mux.NewRouter()
	// Serving static files
	fs := http.FileServer(http.Dir("assets"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))
	r.Use(handlers.LoggingMiddleWare)
	r.HandleFunc("/search", handlers.SearchHandler)
	r.HandleFunc("/", handlers.IndexHandler)
	log.Println("Starting server on : ", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
