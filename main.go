package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
)

var tpl = template.Must(template.ParseFiles("index.html"))

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

	fmt.Println("Search query is: ", searchKey)
	fmt.Println("Results page is: ", page)
}

func main() {
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
