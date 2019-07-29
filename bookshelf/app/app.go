package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"cloud.google.com/go/storage"

	uuid "github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/tony-yang/google-cloud-stack/bookshelf"
)

var (
	UserProfile *Profile
)

// listHandler displays a list with summaries of books in the database.
func listHandler(w http.ResponseWriter, r *http.Request) {
	UserProfile = profileFromSession(r)

	books, err := bookshelf.DB.ListBooks()
	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		bookResult := "<div><a href='login?redirect=books'>Login</a></div><div><a href='logout?redirect=books'>Logout</a></div>"

		if UserProfile != nil {
			fmt.Println("User profile has something =", UserProfile)
			bookResult += fmt.Sprintf("name = %s", UserProfile.DisplayName)
		}

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
	FORM := `<form method="post" enctype="multipart/form-data" action="/books">
		<div class="form-group">
			<label for="title">Title</label>
			<input class="form-control" name="title" id="title">
		</div>
		<div class="form-group">
			<label for="author">Author</label>
			<input class="form-control" name="author" id="author">
		</div>
		<div class="form-group">
			<label for="image">Cover Image</label>
			<input class="form-control" name="image" id="image" type="file">
		</div>
		<input type="submit" name="submit" id="submit" value="Submit">
	</form>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, FORM)
}

func uploadCover(r *http.Request) (url string, err error) {
	f, fileHeader, err := r.FormFile("image")
	if err == http.ErrMissingFile {
		fmt.Println("app.go: uploadCover: Missing File")
		return "", nil
	} else if err != nil {
		fmt.Printf("app.go: uploadCover: %v\n", err)
		return "", err
	}

	if bookshelf.StorageBucket == nil {
		fmt.Println("app.go: uploadCover storage bucket is missing")
		return "", errors.New("storage bucket is missing - check config.go")
	}

	name := uuid.Must(uuid.NewV4()).String() + path.Ext(fileHeader.Filename)
	ctx := context.Background()
	w := bookshelf.StorageBucket.Object(name).NewWriter(ctx)

	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = fileHeader.Header.Get("Content-Type")

	w.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(w, f); err != nil {
		fmt.Printf("app.go: uploadCover: io copy failed %v\n", err)
		return "", err
	}
	if err := w.Close(); err != nil {
		fmt.Printf("app.go: uploadCover: close failed %v\n", err)
		return "", err
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, bookshelf.StorageBucketName, name), nil
}

// createHandler adds a book to the database
func createHandler(w http.ResponseWriter, r *http.Request) {
	book := &bookshelf.Book{
		Title:  r.FormValue("title"),
		Author: r.FormValue("author"),
	}

	imageURL, err := uploadCover(r)
	if err != nil {
		fmt.Printf("createHandler failed to upload cover: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books/add"), http.StatusFound)
	}
	fmt.Println("createHandler: image URL =", imageURL)
	book.ImageURL = imageURL

	id, err := bookshelf.DB.AddBook(book)
	if err != nil {
		fmt.Printf("createHandler failed to add book: %v\n", err)
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

	FORM := fmt.Sprintf(`<form method="post" enctype="multipart/form-data" action="/books/%d">
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
		<div class="form-group">
			<label for="image">Cover Image</label>
			<input class="form-control" name="image" id="image" type="file">
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

	// For OAuth2
	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/oauth2callback", oauthCallbackHandler).Methods("GET")

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