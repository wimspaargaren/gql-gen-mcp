package resolvers

import (
	"context"
	"fmt"

	models1 "github.com/wimspaargaren/gql-gen-mcp/example/bookstore-api/schema/models"
	"github.com/wimspaargaren/gql-gen-mcp/example/bookstore-api/schema/repository"
)

func bookAndAuthorToModel(book *repository.Book, author *repository.Author) *models1.Book {
	return &models1.Book{
		ID:            book.ID,
		Title:         book.Title,
		Description:   &book.Description,
		PublishedYear: &book.PublishedYear,
		Genre:         models1.Genre(book.Genre),
		Price:         book.Price,
		Status:        models1.BookStatus(book.Status),
		Author: &models1.Author{
			ID:        author.ID,
			Name:      author.Name,
			Biography: &author.BioGraphy,
		},
	}
}

func (r *queryResolver) resultingBooksForAuthor(ctx context.Context, id string) ([]*models1.Book, error) {
	books, _, err := r.bookRepository.GetBooks(ctx, repository.BookListInput{
		Filter: &repository.BookFilterInput{
			AuthorID: &id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}
	resultBooks := []*models1.Book{}
	for _, book := range books {
		resultBooks = append(resultBooks, &models1.Book{
			ID:            book.ID,
			Title:         book.Title,
			Description:   &book.Description,
			PublishedYear: &book.PublishedYear,
			Genre:         models1.Genre(book.Genre),
			Price:         book.Price,
			Status:        models1.BookStatus(book.Status),
		})
	}
	return resultBooks, nil
}
