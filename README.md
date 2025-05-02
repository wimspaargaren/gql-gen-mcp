# Overview

The gql-gen-mcp tool has its name inspired by the awesome [github.com/99designs/gqlgen](https://github.com/99designs/gqlgen) generator. Instead of generating a GraphQL server, it generates Model Context Protocol (MCP) (stdio) servers based on your GraphQL schema definitions. It's currently based on [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) as it's the most widely adopted MCP Go server (April 2025). Gotools just started an implementation see [mcp go thread](https://github.com/orgs/modelcontextprotocol/discussions/224#discussioncomment-12924467), so perhaps we eventually switch over or make it configurable which server to use as output.

# Usage

## Install gql-gen-mcp

```bash
go install github.com/wimspaargaren/gql-gen-mcp@latest
```

## Configuration

You can configure gql gen mcp using a yaml file called `.gql-gen-mcp.yaml`.
Example:
```yaml
schemas:
  - name: bookstore
    dir: ./bookstore/graphql/definitions
    output: ./mcp/bookstore
```

## Run

Run the gql-gen-mcp tool in the directory where you've defined your `.gql-gen-mcp.yaml` file. Note that the `main.go` of your server is only generated once, such that you can configure the server to your needs.

```bash
gql-gen-mcp
```

## Use with your favourite LLM tooling

The generated MCP server is a stdio server. Install it on your system with `go install .`.

MCP server definition:
```JSON
"bookstore-api": {
      "command": "bookstore",
      "args": []
  },
```

## Example 

Check out the [example](./example/README.md) directory for a full example with a dummy bookstore GraphQL API.

# Missing features

- Configuration via flags
- No support for introspection or type query
- No support for federated GraphQL
- Generate MCP tools based on introspection query
