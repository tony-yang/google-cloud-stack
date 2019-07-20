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

func detailHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}
	book, err := bookshelf.DB.GetBook(id)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	fmt.Fprintf(w, book.String())
}

// addBookHandler displays a form that captures details of a new book to add.
func addBookHandler(w http.ResponseWriter, r *http.Request) {
	FORM := `<form method="post" action="/books">
		<div class="form-group">
			<label for="title">Title</label>
			<input class="form-control" name="title" id="title">
		</div>
		<div class="form-group">
			<label for="author">Author</label>
			<input class="form-control" name="author" id="author">
		</div>
		<input type="submit" name="submit" id="submit" value="Submit">
	</form>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, FORM)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	book := &bookshelf.Book{
		Title:  r.FormValue("title"),
		Author: r.FormValue("author"),
	}
	id, err := bookshelf.DB.AddBook(book)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%d", id), http.StatusFound)
}

func registerHandlers() {
	fmt.Println("Register handlers")
	r := mux.NewRouter()
	r.HandleFunc("/", listHandler).Methods("GET")
	r.HandleFunc("/books", listHandler).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}", detailHandler).Methods("GET")
	r.HandleFunc("/books/add", addBookHandler).Methods("GET")

	r.HandleFunc("/books", createHandler).Methods("POST")

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
