package klient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_Request(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check request method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request method"}`))
			return
		}

		// check request path
		if r.URL.Path != "/api/v1/test" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request path"}`))
			return
		}

		// check request header
		if r.Header.Get("X-Info") != "test" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request header"}`))
			return
		}

		// get request body
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request body"}`))
			return
		}

		// check request body
		if m["id"] != "123" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid id"}`))
			return
		}

		// write response
		w.Header().Add("X-Info", "test-server")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"request_id": "123+"}`))
	}))

	defer httpServer.Close()

	httpxClient, err := New(
		OptionClient.WithBaseURL(httpServer.URL),
	)
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
		return
	}

	type fields struct {
		HttpClient *http.Client
		BaseURL    *url.URL
	}
	type args struct {
		ctx    context.Context
		method string
		path   string
		body   io.Reader
		header http.Header
		fn     func(*http.Response) error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		long    bool
	}{
		{
			name: "test",
			fields: fields{
				HttpClient: httpxClient.HTTPClient,
				BaseURL:    httpxClient.BaseURL,
			},
			args: args{
				ctx:    context.Background(),
				method: http.MethodPost,
				path:   "/api/v1/test",
				body:   bytes.NewBufferString(`{"id": "123"}`),
				header: http.Header{
					"X-Info": []string{"test"},
				},
				fn: func(resp *http.Response) error {
					// check response status code
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("invalid status code")
					}

					// check response header
					if resp.Header.Get("X-Info") != "test-server" {
						return fmt.Errorf("invalid response header")
					}

					// get response body
					var m map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
						return err
					}

					if m["request_id"] != "123+" {
						return fmt.Errorf("invalid request id")
					}

					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.long && testing.Short() {
				t.Skip("skipping test in short mode.")
			}

			c := &Client{
				HTTPClient: tt.fields.HttpClient,
				BaseURL:    tt.fields.BaseURL,
			}

			if err := c.Request(tt.args.ctx, tt.args.method, tt.args.path, tt.args.body, tt.args.header, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("Client.Request() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
