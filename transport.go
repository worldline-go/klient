package klient

import (
	"context"
	"maps"
	"net/http"
	"net/url"
)

// TransportKlient is an http.RoundTripper that
// wrapping a base RoundTripper and adding headers from context.
type TransportKlient struct {
	// Base is the base RoundTripper used to make HTTP requests.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper
	// Header for default header to set if not exist.
	Header http.Header
	// BaseURL is the base URL for relative requests.
	BaseURL *url.URL

	// Inject extra content to request (e.g. tracing propagation).
	Inject func(ctx context.Context, req *http.Request)
}

var _ http.RoundTripper = (*TransportKlient)(nil)

type ctxKlient string

// TransportHeaderKey is the context key to use with context.WithValue to
// specify http.Header for a request.
var TransportHeaderKey ctxKlient = "HTTP_HEADER"

// RoundTrip authorizes and authenticates the request with an
// access token from Transport's Source.
func (t *TransportKlient) SetHeader(req *http.Request) {
	// add base url
	if t.BaseURL != nil {
		req.URL = t.BaseURL.ResolveReference(req.URL)
	}

	defer func() {
		for k, v := range t.Header {
			if _, ok := req.Header[k]; !ok {
				req.Header[k] = v
			}
		}
	}()

	ctx := req.Context()
	if ctx == nil {
		return
	}

	if header, ok := ctx.Value(TransportHeaderKey).(http.Header); ok {
		maps.Copy(req.Header, header)
	}

	if t.Inject != nil {
		t.Inject(ctx, req)
	}
}

// RoundTrip authorizes and authenticates the request with an
// access token from Transport's Source.
func (t *TransportKlient) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := cloneRequest(req) // per RoundTripper contract
	t.SetHeader(req2)

	return t.base().RoundTrip(req2)
}

func (t *TransportKlient) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}

	return http.DefaultTransport
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}

	return r2
}
