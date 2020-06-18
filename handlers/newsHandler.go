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

// Results is used to recieve response from the api
type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Article defines single article, slice of Articles is used in Results.
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

// FormatPublishedDate is used to format the resp date.
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

// Search is used to get the request param and collect data from the api.
type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

// NewsAPIError will collect data in case api fails.
type NewsAPIError struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

var apiKey *string

// init is used to initialize the flag to send the key. This is the first func that is executed.
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

// IndexHandler is the  default hanlder to execute the template
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	searchKey := "in"
	next := 1
	search := &Search{}
	search.NextPage = next
	pageSize := 20
	endPoint := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=%s&pageSize=%d&page=%d&sortBy=publishedAt&apiKey=%s&language=en", searchKey, pageSize, next, *apiKey)
	resp, err := http.Get(endPoint)
	if err != nil {
		log.Println("Error while calling the endPoint : ", endPoint)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		tpl.Execute(w, nil)
		return
	}
	parseResultIntoTemp(resp, search, w, pageSize)
}

func parseResultIntoTemp(resp *http.Response, search *Search, w http.ResponseWriter, pageSize int) {
	err := json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		log.Println("Error while decoding resp.Body")
		tpl.Execute(w, nil)
		return
	}
	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}
	err = tpl.Execute(w, search)
	if err != nil {
		log.Println("Error while executing the template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// SearchHandler ...
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Get page and q as query params
	query := r.URL.Query()
	searchKey := query.Get("q")
	page := query.Get("page")
	if page == "" {
		page = "1"
	}
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
	// Create the end point and call the service.
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
		newsError := &NewsAPIError{}
		err := json.NewDecoder(resp.Body).Decode(newsError)
		if err != nil {
			http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		}
		http.Error(w, newsError.Message, http.StatusInternalServerError)
		log.Println("Status code is != 200")
		return
	}
	// Decode the response in &search.Results
	parseResultIntoTemp(resp, search, w, pageSize)
}

// LoggingMiddleWare is used to print the elapsed time between a request can and server response.
func LoggingMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(w, r)
		endTime := time.Now()
		elapsed := endTime.Sub(startTime)
		log.Println("Elapsed Time : ", elapsed)
	})
}
