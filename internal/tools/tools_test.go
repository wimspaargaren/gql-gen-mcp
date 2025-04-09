package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestSimpleQuery(t *testing.T) {
	t.Parallel()

	gqlSchma := `
"Query root."
type Query {
  "Retrieves a list of examples."
  example(
    "some string arg."
    string: String
    "some id."
    id: ID
    "some int."
    int: Int
	"some bool."
	bool: Boolean
  ): String!
  }`

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	assert.Equal(t, "example", tools[0].Name)
	assert.Equal(t, "Retrieves a list of examples.", tools[0].Description)
	assert.Equal(t, 4, len(tools[0].Args))
	assert.Equal(t, "some string arg.", tools[0].Args[0].Description)
	assert.Equal(t, "some id.", tools[0].Args[1].Description)
	assert.Equal(t, "some int.", tools[0].Args[2].Description)
	assert.Equal(t, "some bool.", tools[0].Args[3].Description)
	assert.Equal(t, "string", tools[0].Args[0].Name)
	assert.Equal(t, "id", tools[0].Args[1].Name)
	assert.Equal(t, "int", tools[0].Args[2].Name)
	assert.Equal(t, "bool", tools[0].Args[3].Name)
	assert.Equal(t, TypeString, tools[0].Args[0].Type)
	assert.Equal(t, TypeString, tools[0].Args[1].Type)
	assert.Equal(t, TypeNumber, tools[0].Args[2].Type)
	assert.Equal(t, TypeBoolean, tools[0].Args[3].Type)
}

func TestQueryWithInputObject(t *testing.T) {
	t.Parallel()

	gqlSchma := `
	"Filter for listing examples."
input Filter {
  "Filter by ID."
  id: ID
}

"Query root."
type Query {
  "Retrieves a list of examples."
  example(
    "a filter."
    filter: Filter
  ): String!
  }`
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	assert.Equal(t, "example", tools[0].Name)
	assert.Equal(t, "Retrieves a list of examples.", tools[0].Description)
	assert.Equal(t, 1, len(tools[0].Args))
	assert.Equal(t, "a filter.", tools[0].Args[0].Description)
	assert.Equal(t, "filter", tools[0].Args[0].Name)
	assert.Equal(t, TypeObject, tools[0].Args[0].Type)
	assert.Equal(t, `map[string]any{"id": map[string]any{"type": "string", "description": "Filter by ID."}}`, tools[0].Args[0].Properties)
}

func TestQueryWithInputObjectArray(t *testing.T) {
	t.Parallel()

	gqlSchma := `
	"Book object."
input Book {
  "Title of the book."
  title: String!
  "Author of the book."
  author: [String]!
  "Year of publication."
  year: Int
}

"Query root."
type Query {
  "Retrieves a list of examples."
  example(
    "a filter."
    filter: [Book]!
  ): String!
  }`

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	assert.Equal(t, TypeArray, tools[0].Args[0].Type)
	assert.Equal(t,
		`map[string]any{"type": "object", "properties": map[string]any{"title": map[string]any{"type": "string", "description": "Title of the book."}, "author": map[string]any{"type": "array", "description": "Author of the book.", "items": map[string]any{"type": "string"}}, "year": map[string]any{"type": "number", "description": "Year of publication."}}}`,
		tools[0].Args[0].Items)
}

func TestEnumInInputObject(t *testing.T) {
	t.Parallel()

	gqlSchma := `
"Enum representing the family of a tag."
enum BookType {
  "Hardcover book."
  Hardcover
  "Paperback book."
  Paperback
}
input Book {
  "Type of the book."
  type: BookType!
}

"Query root."
type Query {
  "Retrieves a list of examples."
  example(
    "a filter."
    filter: Book
  ): String!
  }
  `
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	assert.Equal(t, TypeObject, tools[0].Args[0].Type)
	assert.Equal(t,
		`map[string]any{"type": map[string]any{"type": "string", "description": "Type of the book.", "enum": []string{"Hardcover", "Paperback"}}}`,
		tools[0].Args[0].Properties)
}

func TestEnum(t *testing.T) {
	t.Parallel()

	gqlSchma := `
"Enum representing the family of a tag."
enum BookType {
  "Hardcover book."
  Hardcover
  "Paperback book."
  Paperback
}

"Query root."
type Query {
  "Retrieves a list of examples."
  example(
    "a filter."
    filter: BookType
  ): String!
  }
  `
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	assert.Equal(t, TypeEnum, tools[0].Args[0].Type)
	assert.Equal(t, []string{"Hardcover", "Paperback"}, tools[0].Args[0].Enum)
}

func TestUnion(t *testing.T) {
	t.Parallel()

	gqlSchma := `
	type Human {
		id: ID!
		name: String!
		totalCredits: Int
	}
	
	type Droid {
		id: ID!
		name: String!
		primaryFunction: String
	}

	"SearchResult is a union type that can be either a Human, Droid, or Starship."
	union SearchResult = Human | Droid

	"Query root."
	type Query {
	"Retrieves a list of examples."
	example(
		"a filter."
		filter: String
	): SearchResult!
  }`

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)

	assert.Equal(t, 1, len(tools))
	compareQueries(t,
		`query example ($filter: String) {
			example(filter: $filter) {
					__typename
					... on Human {
							id 
							name 
							totalCredits 
					}
					... on Droid {
							id 
							name 
							primaryFunction 
					}
			}

}`, tools[0].Query)
}

func TestInterface(t *testing.T) {
	t.Parallel()

	gqlSchema := `
interface Character {
  id: ID!
  name: String!
}

type Human implements Character {
  id: ID!
  name: String!
  totalCredits: Int
}
 
type Droid implements Character {
  id: ID!
  name: String!
  primaryFunction: String
}

"Query root."
	type Query {
	"Retrieves a list of characters."
	example(
		"a filter."
		filter: String
	): Character!
}
  `

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchema,
	})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	compareQueries(t,
		`query example ($filter: String) {
                        example(filter: $filter) {
                                id 
                                name 
                        }

                }`, tools[0].Query)
}

func TestQueryForCyclicSchema(t *testing.T) {
	t.Parallel()

	gqlSchma := `
	"""
An interface that all entities with an ID implement.
"""
interface Node {
  id: ID!
}

	"""
A single author of books.
"""
type Author implements Node {
  id: ID!
  name: String!
  books: [Book!]!
}

"""
Represents a book in the store.
"""
type Book implements Node {
  id: ID!
  title: String!
  author: Author!
}
  
"Query root."
	type Query {
	"Retrieves a list of characters."
	example(
		"a filter."
		filter: String
	): [Book!]!
}`

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Input: gqlSchma,
	})
	assert.NoError(t, err)
	tools := GetToolsForSchema(schema)
	assert.NotNil(t, tools)
	assert.Equal(t, 1, len(tools))
	compareQueries(t,
		` query example ($filter: String) {
                        example(filter: $filter) {
                                id 
                                title 
                                author {
                                        id 
                                        name 
                                }
                        }

                }`, tools[0].Query)
}

func compareQueries(t *testing.T, expected, actual string) {
	t.Helper()
	expected = strings.ReplaceAll(expected, "\n", "")
	expected = strings.ReplaceAll(expected, "\t", "")
	expected = strings.ReplaceAll(expected, " ", "")

	actual = strings.ReplaceAll(actual, "\n", "")
	actual = strings.ReplaceAll(actual, "\t", "")
	actual = strings.ReplaceAll(actual, " ", "")

	assert.Equal(t, expected, actual)
}
