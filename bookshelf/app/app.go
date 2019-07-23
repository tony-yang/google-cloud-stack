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

// listHandler displays a list with summaries of books in the database.
func listHandler(w http.ResponseWriter, r *http.Request) {
	books, err := bookshelf.DB.ListBooks()
	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		bookResult := ""
		for _, book := range books {
			deleteForm := fmt.Sprintf("<form method='post' action='/books/%d/delete'><input type='submit' value='Delete'></form>", book.ID)
			bookResult = bookResult + "<br>" + book.String() + deleteForm
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, string(bookResult))
	}
}

// detailHandler displays the details of a given book.
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

// createHandler adds a book to the database
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

func editHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	book, err := bookshelf.DB.GetBook(id)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	FORM := fmt.Sprintf(`<form method="post" action="/books/%d">
		<div class="form-group">
			<label for="id">ID</label>
			<input class="form-control" name="id" id="id" value="%d">
		</div>
		<div class="form-group">
			<label for="title">Title</label>
			<input class="form-control" name="title" id="title" value="%s">
		</div>
		<div class="form-group">
			<label for="author">Author</label>
			<input class="form-control" name="author" id="author" value="%s">
		</div>
		<input type="submit" name="submit" id="submit" value="Submit">
		</form>`, book.ID, book.ID, book.Title, book.Author)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, FORM)
}

// updateHandler updates a given book with id
func updateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		fmt.Println("update handler parse id error: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}
	book := &bookshelf.Book{
		ID:     id,
		Title:  r.FormValue("title"),
		Author: r.FormValue("author"),
	}
	err = bookshelf.DB.UpdateBook(book)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%d", id), http.StatusFound)
}

// deletHandler deletes a given book
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		fmt.Println("delete handler parse id error: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}
	bookshelf.DB.DeleteBook(id)
	http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
}

func registerHandlers() {
	fmt.Println("Register handlers")
	r := mux.NewRouter()
	r.HandleFunc("/", listHandler).Methods("GET")
	r.HandleFunc("/books", listHandler).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}", detailHandler).Methods("GET")
	r.HandleFunc("/books/add", addBookHandler).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}/edit", editHandler).Methods("GET")

	r.HandleFunc("/books", createHandler).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}", updateHandler).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}/delete", deleteHandler).Methods("POST")

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