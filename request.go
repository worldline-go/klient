package klient

import (
	"context"
	"fmt"
	"net/http"
)

// Requester is the base interface to send an HTTP request.
type Requester[T any] interface {
	// Request returns the http.Request.
	Request(context.Context) (*http.Request, error)
	Response(resp *http.Response) (T, error)
}

// DoWithInf sends an HTTP request and calls the response function with using http.Client and Requester.
func DoWithInf[T any](ctx context.Context, client *http.Client, r Requester[T]) (T, error) {
	var empty T
	if r == nil {
		return empty, ErrRequesterNil
	}

	req, err := r.Request(ctx)
	if err != nil {
		return empty, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	httpResp, err := client.Do(req)
	if err != nil {
		return empty, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	defer DrainBody(httpResp.Body)

	return r.Response(httpResp)
}

// Do sends an HTTP request and calls the response function with using http.Client.
func Do(c *http.Client, req *http.Request, fn func(*http.Response) error) error {
	httpResp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}

	defer DrainBody(httpResp.Body)

	if fn == nil {
		return ErrResponseFuncNil
	}

	return fn(httpResp)
}

// Do sends an HTTP request and calls the response function with resolving reference URL.
//
// It is automatically drain and close the response body.
func (c *Client) Do(req *http.Request, fn func(*http.Response) error) error {
	return Do(c.HTTP, req, fn)
}
