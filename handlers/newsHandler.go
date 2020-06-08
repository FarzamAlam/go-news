package handlers

import (
	"log"
	"net/http"
	"text/template"
)

// index.html is Parsed and if it throws the error then code panics
var tpl = template.Must(template.ParseFiles("index.html"))

// IndexHandler is the  default hanlder to execute the template
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

// SearchHandler ...
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	searchKey := query.Get("q")
	page := query.Get("page")
	if page == "" {
		page = "1"
	}
	log.Println("Search Query is : ", searchKey)
	log.Println("Result Page is : ", page)
}
