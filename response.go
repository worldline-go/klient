package klient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var ResponseErrLimit int64 = 1 << 20 // 1MB

// ResponseFuncJSON returns a response function that decodes the response into data
// with json decoder. It will return an error if the response status code is not 2xx.
//
// If data is nil, it will not decode the response body.
func ResponseFuncJSON(data interface{}) func(*http.Response) error {
	return func(resp *http.Response) error {
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
}

// LimitedResponse not close body, retry library draining it.
//   - Return limited response body
//   - Ready all body and assign it back to resp.Body
func LimitedResponse(resp *http.Response) []byte {
	if resp == nil {
		return nil
	}

	v, _ := io.ReadAll(io.LimitReader(resp.Body, ResponseErrLimit))

	resp.Body = NewMultiReader(io.NopCloser(bytes.NewReader(v)), resp.Body)

	return v
}
