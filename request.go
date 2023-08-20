package klient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type request struct {
	method string
	path   string
	header http.Header
	body   io.Reader
}

func (r request) Method() string {
	return r.method
}

func (r request) Path() string {
	return r.path
}

func (r request) Header() http.Header {
	return r.header
}

func (r request) Body() io.Reader {
	return r.body
}

var globalClient, _ = NewClient(OptionClient.WithDisableBaseURLCheck(true))

// Request sends an HTTP request and calls the response function with the global client.
func Request(ctx context.Context, baseURL, method, path string, body io.Reader, header http.Header, fn func(*http.Response) error) error {
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	client := globalClient
	client.BaseURL = baseURLParsed

	return client.DoWithFunc(ctx, request{
		method: method,
		path:   path,
		header: header,
		body:   body,
	}, fn)
}

// Request sends an HTTP request and calls the response function.
func (c *Client) Request(ctx context.Context, method, path string, body io.Reader, header http.Header, fn func(*http.Response) error) error {
	return c.DoWithFunc(ctx, request{
		method: method,
		path:   path,
		header: header,
		body:   body,
	}, fn)
}
