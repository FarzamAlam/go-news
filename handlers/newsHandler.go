package handlers

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	gocache "github.com/patrickmn/go-cache"
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
	return a.PublishedAt.Format("January 02, 2006")
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

// IsLastPage returns the true if the next page is equal to total page
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

// CurrentPage returns the current page.s
func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}
	return s.NextPage - 1
}

// PreviousPage returns current page - 1s
func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

// cache variable is used to cache the search object, it expires after 15 min.
var cache = gocache.New(15*time.Minute, 20*time.Minute)

// IndexHandler is the  default hanlder to execute the template
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Our Web app by default shows the top 20 results for country = India(in)
	searchKey := "in"
	next := 1
	search := &Search{}
	search.NextPage = next
	pageSize := 20
	endPoint := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=%s&pageSize=%d&page=%d&sortBy=publishedAt&apiKey=%s&language=en", searchKey, pageSize, next, *apiKey)
	// getAPIData is called with search object, pageSize, endPoint and response writer object
	// it used to get the api data and parse it into the template.
	getAPIData(search, pageSize, endPoint, w)
}

// parseResultIntoTemp takes the searchObject calculates the total pages and increases the next page
// and parses the search object into the template.
func parseResultIntoTemp(search *Search, w http.ResponseWriter, pageSize int) {
	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}
	err := tpl.Execute(w, search)
	if err != nil {
		log.Println("Error while executing the template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// SearchHandler ... It is used to handle the search query.
// It takes two query params q and page. where q is search string and page is the page number.
// It parses the page from the query param creates the endPoint with the query string and calls the getAPIData
// to get and parse the result on template.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Get page and q as query params
	query := r.URL.Query()
	searchKey := query.Get("q")
	// If the query is empty then return index page.
	if searchKey == "" {
		IndexHandler(w, r)
		return
	}
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
	getAPIData(search, pageSize, endPoint, w)
}

// getAPIData is used to get the search object from the cache or from the api call.
// It first checks the cache with the endPoint, if it finds then it calls the parseResultIntoTemp func and returns.
// It endPoint is not present in the cache memory then it calls the api with the endPoint.
// If it results in the error then it returns the error in NewsAPIError object.
// Otherwise it decodes the reponse object in the search and calls parseResultIntoTemp func.
// and at last it caches the search object and endpoint string.
func getAPIData(search *Search, pageSize int, endPoint string, w http.ResponseWriter) {
	if search, ok := cache.Get(endPoint); ok {
		parseResultIntoTemp(search.(*Search), w, pageSize)
		return
	}
	// Calling the external API.
	resp, err := http.Get(endPoint)
	if err != nil {
		log.Println("Error calling the endPoint : ", err)
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
	// decoding the response body into &search.Results
	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		log.Println("Error while decoding resp.Body : ", err)
		tpl.Execute(w, nil)
		return
	}
	// Parse the response in template
	parseResultIntoTemp(search, w, pageSize)
	cache.Set(endPoint, search, gocache.DefaultExpiration)
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
