package handlers

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"
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

	search := &Search{}
	search.SearchKey = searchKey
	next, err := strconv.Atoi(page)
	if err != nil {
		log.Println("Error while parsing page :", err)
		http.Error(w, "Unexpected Server error", http.StatusInternalServerError)
		return
	}
	search.NextPage = next
	pageSize := 20
	endPoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
	resp, err := http.Get(endPoint)
	if err != nil {
		log.Println("Error calling the endPoint")
		log.Println("endPoint : ", endPoint)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("Status code is != 200")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		log.Println("Error while decoding the json body : ", err)
	}
	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}
	err = tpl.Execute(w, search)
	if err != nil {
		log.Println("Error while tpl.Execute : ", err)
	}
}

type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

var apiKey *string

func init() {
	apiKey = flag.String("apiKey", "", "Newsapi.org access key.")
	flag.Parse()
	if *apiKey == "" {
		log.Fatal("apiKey must be set.")
	}
}

func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}
	return s.NextPage - 1
}

func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}
