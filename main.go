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
// 4. Implement Caching.		--> Done.
// 5. Write Comments			--> Done.
// 6. Deploy on heruko.			--> Done.
// 7. Make a proper README.
// 8. Post it on reddit.

func main() {
	// Getting the port from the environment variable.
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	r := mux.NewRouter()
	// Serving static files
	fs := http.FileServer(http.Dir("assets"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))

	// Using a middleware to print the execution time of each request.
	r.Use(handlers.LoggingMiddleWare)

	// Handlers to handle the search result or index page.
	r.HandleFunc("/search", handlers.SearchHandler)
	r.HandleFunc("/", handlers.IndexHandler)

	// Starting the server
	log.Println("Starting server on : ", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
