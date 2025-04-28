// Package main bootstraps the MCP server and registers the tools.
package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"

	"github.com/wimspaargaren/gql-gen-mcp/graphql"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your server
// add any graphql client and MCP server configurations here.

func main() {
	s := server.NewMCPServer(
		"GraphQL MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	gqlClient := graphql.NewDefaultClient("http://127.0.0.1:8080/query")

	toolRegistry := NewToolRegistry(s, gqlClient)
	toolRegistry.RegisterTools()

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %s\n", err)
	}
}
