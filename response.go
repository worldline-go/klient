package klient

import (
	"io"
	"net/http"
)

var ResponseErrLimit int64 = 1 << 20 // 1MB

// LimitedResponse not close body, retry library draining it.
func LimitedResponse(resp *http.Response) []byte {
	v, _ := io.ReadAll(io.LimitReader(resp.Body, ResponseErrLimit))

	return v
}

func drainBody(body io.ReadCloser) {
	defer body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(body, ResponseErrLimit))
}
