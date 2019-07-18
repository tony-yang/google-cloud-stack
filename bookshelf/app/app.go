package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tony-yang/google-cloud-stack/bookshelf"
)

func listHandler(w http.ResponseWriter, r *http.Request) {
	books, _ := bookshelf.DB.ListBooks()

	bookResults := ""
	for _, book := range books {
		bookResults = bookResults + strconv.FormatInt(book.ID, 10) + ": " + book.Title + " by " + book.Author
	}
	fmt.Fprintf(w, bookResults)
}

func registerHandlers() {
	fmt.Println("Register handlers")
	r := mux.NewRouter()
	r.HandleFunc("/", listHandler).Methods("GET")
	r.HandleFunc("/books", listHandler).Methods("GET")
	
	http.Handle("/", r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	fmt.Println("Starting the server on port:", port)
	registerHandlers()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}