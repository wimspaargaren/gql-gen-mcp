# Bookstore API Example

This example shows you a dummy bookstore with a GraphQL API server with some LLM hallucinated books and authors. You can list, get, update, create and delete books and list and get authors.

# Run on your machine

1. `go run ./bookstore-api/cmd/bookstore-api`
2. `go install mcp/bookstore`
3. `go install github.com/mark3labs/mcphost@latest`
4. Configure `~/.msp.json`
```JSON
{
  "mcpServers": {
    "bookstore-api": {
      "command": "bookstore",
      "args": []
    }
  }
}
```
5. Run it locally
  With ollama `mcphost -m ollama:llama3.1:70b`
  With openai `mcphost -m openai:gpt-4`
  (stronger models give better results in general)
