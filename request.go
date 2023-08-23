package klient

import (
	"fmt"
	"net/http"
)

// Do sends an HTTP request and calls the response function with resolving reference URL.
//
// It is automatically drain and close the response body.
func (c *Client) Do(req *http.Request, fn func(*http.Response) error) error {
	req.URL = c.BaseURL.ResolveReference(req.URL)

	httpResp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}

	defer drainBody(httpResp.Body)

	if fn == nil {
		return ErrResponseFuncNil
	}

	return fn(httpResp)
}
