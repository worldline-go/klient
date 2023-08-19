package klient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// InfRequest is the base interface for all requests.
type InfRequest interface {
	// Method returns the HTTP method.
	Method() string
	// Path returns the path to be sent.
	Path() string
}

type InfHeader interface {
	// Header returns the header to be sent.
	Header() http.Header
}

type InfRequestValidator interface {
	// Validate returns error if request is invalid.
	Validate() error
}

type InfQueryStringGenerator interface {
	// ToQuery returns the query string to be sent.
	ToQuery() url.Values
}

type InfBodyJSON interface {
	// BodyJSON can return any type that can be marshaled to JSON.
	// Automatically sets Content-Type to application/json.
	BodyJSON() interface{}
}

type InfBody interface {
	// Body returns the body to be sent.
	Body() io.Reader
}

type optionDoValue struct {
	baseURL *url.URL
	header  http.Header
}

// OptionClient is a function that configures the client.
type optionDoFn func(*optionDoValue)

type optionDo struct{}

var OptionDo = optionDo{}

func (optionDo) WithBaseURL(baseURL *url.URL) optionDoFn {
	return func(o *optionDoValue) {
		o.baseURL = baseURL
	}
}

func (optionDo) WithHeader(header http.Header) optionDoFn {
	return func(o *optionDoValue) {
		o.header = header
	}
}

// DoWithFunc sends an HTTP request and calls the response function.
//
// Request additional implements InfRequestValidator, InfQueryStringGenerator, InfHeader, InfBody, InfBodyJSON.
func (c *Client) DoWithFunc(ctx context.Context, req InfRequest, fn func(*http.Response) error, opts ...optionDoFn) error {
	var option optionDoValue
	for _, opt := range opts {
		opt(&option)
	}

	baseURL := c.BaseURL
	if option.baseURL != nil {
		baseURL = option.baseURL
	}

	if baseURL == nil {
		return fmt.Errorf("base url is required")
	}

	if v, ok := req.(InfRequestValidator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrValidating, err)
		}
	}

	u := &url.URL{
		Path: req.Path(),
	}

	if g, ok := req.(InfQueryStringGenerator); ok {
		u.RawQuery = g.ToQuery().Encode()
	}

	var (
		header http.Header
		body   io.Reader
	)

	// set default header
	if h, ok := req.(InfHeader); ok {
		header = h.Header()
	}

	if header == nil {
		header = make(http.Header)
	}

	if b, ok := req.(InfBody); ok {
		body = b.Body()
	} else if jb, ok := req.(InfBodyJSON); ok {
		bodyGet := jb.BodyJSON()

		bodyData, err := json.Marshal(bodyGet)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMarshal, err)
		}

		body = bytes.NewReader(bodyData)

		header.Set("Content-Type", "application/json")
	}

	// add context values
	if option.header != nil {
		for k := range option.header {
			header.Set(k, option.header.Get(k))
		}
	}

	uSend := baseURL.ResolveReference(u)

	httpReq, _ := http.NewRequestWithContext(ctx, req.Method(), uSend.String(), body)
	httpReq.Header = header

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRequest, err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, httpResp.Body)
		_ = httpResp.Body.Close()
	}()

	if fn == nil {
		return ErrResponseFuncNil
	}

	return fn(httpResp)
}

// Do sends an HTTP request and json unmarshals the response body to data.
//
// Do work same as DoWithFunc with defaultResponseFunc.
func (c *Client) Do(ctx context.Context, req InfRequest, resp interface{}) error {
	fn := func(r *http.Response) error { return defaultResponseFunc(r, resp) }

	return c.DoWithFunc(ctx, req, fn)
}

func defaultResponseFunc(resp *http.Response, data interface{}) error {
	if err := UnexpectedResponse(resp); err != nil {
		return err
	}

	// 204s, for example
	if resp.ContentLength == 0 {
		return nil
	}

	if data == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}
