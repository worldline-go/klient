package klient

import (
	"context"
	"fmt"
	"net/http"
)

// Requester is the base interface to send an HTTP request.
type Requester interface {
	// Request returns the http.Request.
	Request(context.Context) (*http.Request, error)
}

// DoWithInf sends an HTTP request and calls the response function with resolving reference URL.
//
// It is automatically drain and close the response body.
func (c *Client) DoWithInf(ctx context.Context, request Requester, fn func(*http.Response) error) error {
	req, err := request.Request(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	return c.Do(req, fn)
}

// Do sends an HTTP request and calls the response function with resolving reference URL.
//
// It is automatically drain and close the response body.
func (c *Client) Do(req *http.Request, fn func(*http.Response) error) error {
	httpResp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}

	defer DrainBody(httpResp.Body)

	if fn == nil {
		return ErrResponseFuncNil
	}

	return fn(httpResp)
}
