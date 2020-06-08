package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

// index.html is Parsed and if it throws the error then code panics
var tpl = template.Must(template.ParseFiles("index.html"))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	log.Println("Starting server on : ", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}
