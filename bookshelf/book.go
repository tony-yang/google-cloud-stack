package bookshelf

import (
	"fmt"
)

// Book holds metadata about a book
type Book struct {
	ID       int64
	Title    string
	Author   string
	ImageURL string
}

func (b *Book) String() string {
	return fmt.Sprintf("ID: %d => Title: %s, Author: %s, ImageURL: %s", b.ID, b.Title, b.Author, b.ImageURL)
}