package bookshelf

// BookDatabase provides thread-safe access to a database of books.
type BookDatabase interface {
	// ListBooks returns a list of books, ordered by title
	ListBooks() ([]*Book, error)

	// GetBook retrieves a book by its ID
	GetBook(id int64) (*Book, error)

	// AddBook saves a given book, assigning it a new ID
	AddBook(b *Book) (id int64, err error)

	// DeleteBook removes a given book by its ID
	DeleteBook(id int64) error

	// UpdateBook updates the entry for a given book
	UpdateBook(b *Book) error

	// Close closes the database, freeing up resources
	Close()
}
