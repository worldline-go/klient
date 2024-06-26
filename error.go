package klient

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrCreateRequest   = errors.New("failed to create request")
	ErrRequest         = errors.New("failed to do request")
	ErrResponseFuncNil = errors.New("response function is nil")
	ErrRequesterNil    = errors.New("requester is nil")
)

type ResponseError struct {
	StatusCode int
	Body       string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("unexpected response (%d): %s", e.StatusCode, e.Body)
}

// ErrResponse returns an error with the limited response body.
func ErrResponse(resp *http.Response) error {
	partialBody := LimitedResponse(resp)

	return &ResponseError{
		StatusCode: resp.StatusCode,
		Body:       string(partialBody),
	}
}

// UnexpectedResponse returns an error if the response status code is not 2xx with calling ResponseError.
func UnexpectedResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrResponse(resp)
	}

	return nil
}
