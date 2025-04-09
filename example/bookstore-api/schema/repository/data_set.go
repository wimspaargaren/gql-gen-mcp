package repository

import (
	"log"

	"github.com/google/uuid"
)

// InitialiseDummyData creates a set of dummy data for authors and books.
func InitialiseDummyData() (map[string]*Author, map[string]*Book) { //nolint:funlen
	// Create a map for authors with uuid.UUID as the key.
	authors := make(map[string]*Author)

	// Create a few fictional authors.
	authorIDs := []uuid.UUID{}
	authorData := []struct {
		Name      string
		BioGraphy string
	}{
		{"Jane Doe", "Jane Doe is an award-winning novelist known for her compelling storytelling."},
		{"John Smith", "John Smith writes thrilling science fiction and has a background in astrophysics."},
		{"Alice Johnson", "Alice Johnson is a historical fiction author whose vivid characters bring the past to life."},
	}

	// Populate the authors map.
	for _, data := range authorData {
		id := uuid.New()
		authorIDs = append(authorIDs, id)

		authors[id.String()] = &Author{
			ID:        id.String(),
			Name:      data.Name,
			BioGraphy: data.BioGraphy,
		}
	}

	// Print authors map to verify
	log.Default().Println("Authors:")
	for id, author := range authors {
		log.Default().Printf("ID: %v, Author: %+v\n", id, author)
	}

	// Create a map for books with uuid.UUID as the key.
	books := make(map[string]*Book)

	// Create a few fictional books linking to the existing authors.
	// For simplicity, we'll assign books in a round-robin fashion to the created authors.
	bookData := []struct {
		Title         string
		Description   string
		PublishedYear int
		Genre         Genre
		Price         float64
		Status        BookStatus
	}{
		{
			Title:         "The Mystery of the Lost City",
			Description:   "A gripping mystery that takes readers on a journey through an ancient civilization.",
			PublishedYear: 2015,
			Genre:         GenreMystery,
			Price:         19.99,
			Status:        BookStatusAvailable,
		},
		{
			Title:         "Cosmos and Beyond",
			Description:   "An exploration of space and the scientific wonders of our universe.",
			PublishedYear: 2018,
			Genre:         GenreScience,
			Price:         24.50,
			Status:        BookStatusAvailable,
		},
		{
			Title:         "Legends of the Past",
			Description:   "A historical fiction that reimagines stories from a bygone era.",
			PublishedYear: 2020,
			Genre:         GenreHistory,
			Price:         17.75,
			Status:        BookStatusOutOfStock,
		},
		{
			Title:         "Enchanted Realms",
			Description:   "Dive into a world of fantasy filled with magic, dragons, and epic battles.",
			PublishedYear: 2021,
			Genre:         GenreFantasy,
			Price:         22.00,
			Status:        BookStatusAvailable,
		},
	}

	// Assign each book an author in a round-robin manner.
	for i, data := range bookData {
		bookID := uuid.New()
		// Link the book with one of the authors from authorIDs.
		authorID := authorIDs[i%len(authorIDs)]
		books[bookID.String()] = &Book{
			ID:            bookID.String(),
			Title:         data.Title,
			Description:   data.Description,
			PublishedYear: data.PublishedYear,
			Genre:         data.Genre,
			Price:         data.Price,
			Status:        data.Status,
			AuthorID:      authorID.String(),
		}
	}

	log.Default().Println("\nBooks:")
	for id, book := range books {
		log.Default().Printf("ID: %v, Book: %+v\n", id, book)
	}
	return authors, books
}
