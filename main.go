package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var tpl = template.Must(template.ParseFiles("index.html"))

var apiKey *string

type Source struct {
	ID	interface{} `json:"id"`
	Name	string	`json:"name"`
}

type Article struct {
	Source	Source	`json:"source"`
	Author	string	`json:"author"`
	Title	string	`json:"title"`
	Description	string	`json:"description"`
	URL	string	`json:"url"`
	URLToImage	string	`json:"urlToImage"`
	PublishedAt	time.Time	`json:"publishedAt"`
	Content	string	`json:"content"`
}

type Results struct {
	Status	string	`json:"status"`
	TotalResults	int	`json:"totalResults"`
	Articles	[]Article	`json:"articles"`
}

type Search struct {
	SearchKey string
	NextPage	int
	TotalPages	int
	Results	Results
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
//	w.Write([]byte("<h1>Hello, World!</h1>"))
	tpl.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("*** Internal Server Error. Unable to parse request URL."))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	//fmt.Println("Search query is: ", searchKey)
	//fmt.Println("Results page is: ", page)
	search := &Search{}
	search.SearchKey = searchKey

	next, err := strconv.Atoi(page)
	if err != nil {
		log.Println(err)
		http.Error(w, "*** Unexpected Server Error. Unable to parse page number.", http.StatusInternalServerError)
		return
	}

	search.NextPage = next
	pageSize := 20

	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("resp.StatusCode: ", resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	err = tpl.Execute(w, search)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	
}

func main() {
	// Read NewsAPI API key from command line
	apiKey = flag.String("apikey", "", "NewsAPI.org API key")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("NewsAPI.org API key must be provided.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}
