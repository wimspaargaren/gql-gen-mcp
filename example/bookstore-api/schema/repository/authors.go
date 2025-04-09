package repository

import (
	"context"
	"fmt"
)

// Author represents an author of a book.
type Author struct {
	ID        string
	Name      string
	BioGraphy string
}

// Authors is an interface that defines methods for managing authors.
type Authors interface {
	// GetAuthor retrieves an author by ID.
	GetAuthor(ctx context.Context, id string) (*Author, error)
	// GetAuthors retrieves all authors.
	GetAuthors(ctx context.Context) ([]*Author, error)
}

type authorsMem struct {
	// books is a map of book IDs to books.
	authors map[string]*Author
}

// GetAuthor retrieves an author by ID.
func (a *authorsMem) GetAuthor(_ context.Context, id string) (*Author, error) {
	author, ok := a.authors[id]
	if !ok {
		return nil, fmt.Errorf("author with ID %s not found", id)
	}
	return author, nil
}

// GetAuthors retrieves all authors.
func (a *authorsMem) GetAuthors(_ context.Context) ([]*Author, error) {
	authors := make([]*Author, 0, len(a.authors))
	for _, author := range a.authors {
		authors = append(authors, author)
	}
	return authors, nil
}

// NewMemoryAuthors creates a new in-memory author repository.
func NewMemoryAuthors(authors map[string]*Author) Authors {
	return &authorsMem{
		authors: authors,
	}
}
