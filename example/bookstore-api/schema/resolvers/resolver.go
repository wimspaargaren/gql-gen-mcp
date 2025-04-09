// Package resolvers contains the GraphQL resolvers for the bookstore API.
package resolvers

import (
	"github.com/99designs/gqlgen/graphql"

	"github.com/wimspaargaren/gql-gen-mcp/example/bookstore-api/schema/repository"
	"github.com/wimspaargaren/gql-gen-mcp/example/bookstore-api/schema/server"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver serves as the root resolver for the GraphQL server.
type Resolver struct {
	bookRepository   repository.Books
	authorRepository repository.Authors
}

// NewResolver constructs a new Resolver for the GraphQL server.
func NewResolver() graphql.ExecutableSchema {
	authors, books := repository.InitialiseDummyData()
	resolver := &Resolver{
		bookRepository:   repository.NewMemoryBooks(books),
		authorRepository: repository.NewMemoryAuthors(authors),
	}
	schemaCfg := server.Config{
		Resolvers: resolver,
	}
	return server.NewExecutableSchema(schemaCfg)
}
