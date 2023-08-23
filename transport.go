package klient

import (
	"net/http"
)

// Transport is an http.RoundTripper that
// wrapping a base RoundTripper and adding headers from context.
type TransportCtx struct {
	// Base is the base RoundTripper used to make HTTP requests.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper
}

var _ http.RoundTripper = (*TransportCtx)(nil)

type ctxKlient string

// TransportCtxKey is the context key to use with context.WithValue to
// specify http.Header for a request.
var TransportCtxKey ctxKlient = "HTTP_HEADER"

// RoundTrip authorizes and authenticates the request with an
// access token from Transport's Source.
func (t *TransportCtx) SetHeader(req *http.Request) {
	ctx := req.Context()
	if ctx == nil {
		return
	}

	if header, ok := ctx.Value(TransportCtxKey).(http.Header); ok {
		for k, v := range header {
			req.Header[k] = v
		}
	}
}

// RoundTrip authorizes and authenticates the request with an
// access token from Transport's Source.
func (t *TransportCtx) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := CloneRequest(req) // per RoundTripper contract
	t.SetHeader(req2)

	return t.base().RoundTrip(req2)
}

func (t *TransportCtx) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}

	return http.DefaultTransport
}

// CloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func CloneRequest(r *http.Request) *http.Request {
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
