package klient

import (
	"fmt"
	"net/http"
)

var (
	ErrValidating      = fmt.Errorf("failed to validate request")
	ErrMarshal         = fmt.Errorf("failed to marshal request body")
	ErrRequest         = fmt.Errorf("failed to do request")
	ErrResponseFuncNil = fmt.Errorf("response function is nil")
)

// ResponseError returns an error with the limited response body.
func ResponseError(resp *http.Response) error {
	partialBody := LimitedResponse(resp)

	return fmt.Errorf("unexpected response (%d): %s", resp.StatusCode, string(partialBody))
}

// UnexpectedResponse returns an error if the response status code is not 2xx.
func UnexpectedResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ResponseError(resp)
	}

	return nil
}
