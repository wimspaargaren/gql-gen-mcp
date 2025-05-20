// Package graphql provides a simple GraphQL client for making requests to a GraphQL server.
package graphql

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Request represents a GraphQL request.
type Request struct {
	Query         string `json:"query"`
	Variables     any    `json:"variables,omitempty"`
	OperationName string `json:"operationName,omitempty"`
}

// HTTPRequestHook is a function that can be used to modify the HTTP request before it is sent.
type HTTPRequestHook func(req *http.Request) error

type response struct {
	Data   json.RawMessage `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

// Client is a GraphQL client that can be used to send requests to a GraphQL server.
type Client struct {
	baseURL    string
	httpClient *http.Client
	hooks      []HTTPRequestHook
}

// NewDefaultClient creates a new GraphQL client with the default HTTP client and the provided base URL.
func NewDefaultClient(baseURL string, requestHooks ...HTTPRequestHook) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		hooks:      requestHooks,
	}
}

// Call executes a GraphQL request and decodes the response into the provided result variable.
func (c *Client) Call(ctx context.Context, request Request, result any) error { //nolint:revive
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request struct failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request struct failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	for _, hook := range c.hooks {
		if err := hook(req); err != nil {
			return fmt.Errorf("request hook failed: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Default().Println("close response body failed", err)
		}
	}()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		resp.Body, err = gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("gzip decode failed: %w", err)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: %w", err)
	}

	return c.processResponse(body, result)
}

func (c *Client) processResponse(body []byte, result any) error {
	gqlResponse := response{}
	err := json.Unmarshal(body, &gqlResponse)
	if err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}
	if len(gqlResponse.Errors) > 0 {
		b, err := json.Marshal(gqlResponse.Errors)
		if err != nil {
			return fmt.Errorf("marshal error response failed: %w", err)
		}
		return fmt.Errorf("graphql error: %s", string(b))
	}

	err = json.Unmarshal(gqlResponse.Data, result)
	if err != nil {
		return fmt.Errorf("unmarshal data failed: %w", err)
	}
	return nil
}
