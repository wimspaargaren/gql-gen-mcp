// Package repository provides an in-memory implementation of the Books interface.
package repository

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// CreateBookInput is used to create a new book.
type CreateBookInput struct {
	Title         string
	Description   string
	PublishedYear int
	Genre         Genre
	Price         float64
	Status        BookStatus
	AuthorID      string
}

// UpdateBookInput is used to update a book.
type UpdateBookInput struct {
	Title         *string
	Description   *string
	PublishedYear *int
	Genre         *Genre
	Price         *float64
	Status        *BookStatus
	AuthorID      *string
}

// BookListInput is used to list books.
type BookListInput struct {
	Filter    *BookFilterInput
	First     *int
	After     *string
	SortBy    *string
	SortOrder *string
}

// BookFilterInput is used to filter books.
type BookFilterInput struct {
	Genre           *Genre
	Status          *BookStatus
	AuthorID        *string
	MinPrice        *float64
	MaxPrice        *float64
	PublishedAfter  *int
	PublishedBefore *int
	SearchText      *string
}

// Genre the genre of a book.
type Genre string

// Genre definitions.
const (
	GenreFiction    Genre = "FICTION"
	GenreNonFiction Genre = "NON_FICTION"
	GenreScience    Genre = "SCIENCE"
	GenreHistory    Genre = "HISTORY"
	GenreFantasy    Genre = "FANTASY"
	GenreBiography  Genre = "BIOGRAPHY"
	GenreChildren   Genre = "CHILDREN"
	GenreRomance    Genre = "ROMANCE"
	GenreThriller   Genre = "THRILLER"
	GenreMystery    Genre = "MYSTERY"
	GenreSelfHelp   Genre = "SELF_HELP"
)

// BookStatus the status of a book in the store.
type BookStatus string

// BookStatus definitions.
const (
	BookStatusAvailable    BookStatus = "AVAILABLE"
	BookStatusOutOfStock   BookStatus = "OUT_OF_STOCK"
	BookStatusDiscontinued BookStatus = "DISCONTINUED"
)

// Book represents a book in the store.
type Book struct {
	ID            string
	Title         string
	Description   string
	PublishedYear int
	Genre         Genre
	Price         float64
	Status        BookStatus
	AuthorID      string
}

// Books is an interface that defines the methods for managing books in the store.
type Books interface {
	// CreateBook creates a new book.
	CreateBook(ctx context.Context, input *CreateBookInput) (*Book, error)
	// UpdateBook updates an existing book.
	UpdateBook(ctx context.Context, id string, input UpdateBookInput) (*Book, error)
	// DeleteBook deletes a book by ID.
	DeleteBook(ctx context.Context, id string) (bool, error)
	// GetBook retrieves a book by ID.
	GetBook(ctx context.Context, id string) (*Book, error)
	// GetBooks retrieves all books.
	GetBooks(ctx context.Context, input BookListInput) ([]*Book, int, error)
}

type booksMem struct {
	// books is a map of book IDs to books.
	books map[string]*Book
}

// CreateBook creates a new book.
func (b *booksMem) CreateBook(_ context.Context, input *CreateBookInput) (*Book, error) {
	book := &Book{
		ID:            uuid.New().String(),
		Title:         input.Title,
		Description:   input.Description,
		PublishedYear: input.PublishedYear,
		Genre:         input.Genre,
		Price:         input.Price,
		Status:        input.Status,
		AuthorID:      input.AuthorID,
	}
	b.books[book.ID] = book
	return book, nil
}

// UpdateBook updates an existing book.
func (b *booksMem) UpdateBook(_ context.Context, id string, input UpdateBookInput) (*Book, error) { //nolint: revive
	book, ok := b.books[id]
	if !ok {
		return nil, fmt.Errorf("book with ID %s not found", id)
	}
	if input.Title != nil {
		book.Title = *input.Title
	}
	if input.Description != nil {
		book.Description = *input.Description
	}
	if input.PublishedYear != nil {
		book.PublishedYear = *input.PublishedYear
	}
	if input.Genre != nil {
		book.Genre = *input.Genre
	}
	if input.Price != nil {
		book.Price = *input.Price
	}
	if input.Status != nil {
		book.Status = *input.Status
	}
	b.books[book.ID] = book
	return book, nil
}

// DeleteBook deletes a book by ID.
func (b *booksMem) DeleteBook(_ context.Context, id string) (bool, error) {
	book, ok := b.books[id]
	if !ok {
		return false, fmt.Errorf("book with ID %s not found", id)
	}
	delete(b.books, book.ID)
	return true, nil
}

// GetBook retrieves a book by ID.
func (b *booksMem) GetBook(_ context.Context, id string) (*Book, error) {
	book, ok := b.books[id]
	if !ok {
		return nil, fmt.Errorf("book with ID %s not found", id)
	}
	return book, nil
}

// GetBooks retrieves all books.
func (b *booksMem) GetBooks(_ context.Context, input BookListInput) ([]*Book, int, error) { //nolint:revive,gocyclo,gocognit
	books := make([]*Book, 0, len(b.books))
	for _, book := range b.books {
		//nolint:nestif
		if input.Filter != nil {
			if input.Filter.Genre != nil && book.Genre != *input.Filter.Genre {
				continue
			}
			if input.Filter.Status != nil && book.Status != *input.Filter.Status {
				continue
			}
			if input.Filter.AuthorID != nil && book.AuthorID != *input.Filter.AuthorID {
				continue
			}
			if input.Filter.MinPrice != nil && book.Price < *input.Filter.MinPrice {
				continue
			}
			if input.Filter.MaxPrice != nil && book.Price > *input.Filter.MaxPrice {
				continue
			}
			if input.Filter.PublishedAfter != nil && book.PublishedYear < *input.Filter.PublishedAfter {
				continue
			}
			if input.Filter.PublishedBefore != nil && book.PublishedYear > *input.Filter.PublishedBefore {
				continue
			}
		}
		books = append(books, book)
	}
	filterColumn := "id"
	if input.SortBy != nil {
		filterColumn = *input.SortBy
	}
	sortOrder := 1
	if input.SortOrder != nil && *input.SortOrder == "desc" {
		sortOrder = -1
	}
	// order the books
	slices.SortFunc(books, func(a, b *Book) int {
		if filterColumn == "id" && a.ID > b.ID {
			return sortOrder
		}
		if filterColumn == "price" && a.Price > b.Price {
			return sortOrder
		}
		if filterColumn == "publishedYear" && a.PublishedYear > b.PublishedYear {
			return sortOrder
		}
		if filterColumn == "title" && a.Title > b.Title {
			return sortOrder
		}
		return 0
	})
	res := []*Book{}
	for i := 0; i < len(books); i++ {
		if input.First != nil && i >= *input.First {
			break
		}
		res = append(res, books[i])
		// fixme implement after
	}

	return res, len(b.books), nil
}

// NewMemoryBooks creates a new in-memory book repository.
func NewMemoryBooks(books map[string]*Book) Books {
	return &booksMem{
		books: books,
	}
}
