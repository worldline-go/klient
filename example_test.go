package klient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/worldline-go/klient"
)

type CreateXRequest struct {
	ID string `json:"id"`
}

func (r CreateXRequest) Request(ctx context.Context) (*http.Request, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	bodyData, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request body: %w", err)
	}

	body := bytes.NewReader(bodyData)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/x", body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Info", "example")

	return req, nil
}

func (CreateXRequest) Response(resp *http.Response) (CreateXResponse, error) {
	var v CreateXResponse

	if err := klient.UnexpectedResponse(resp); err != nil {
		return v, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, err
	}

	return v, nil
}

type CreateXResponse struct {
	RequestID string `json:"request_id"`
}

// ---

func Example() {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check request method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request method"}`))
			return
		}

		// check request path
		if r.URL.Path != "/api/v1/x" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request path"}`))
			return
		}

		// check request header
		if r.Header.Get("X-Info") != "example" {
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"request_id": "123+"}`))
	}))

	defer httpServer.Close()

	// klient.DefaultBaseURL = httpServer.URL
	os.Setenv("API_GATEWAY_ADDRESS", httpServer.URL)

	client, err := klient.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	v, err := klient.DoWithInf(context.Background(), client.HTTP, CreateXRequest{
		ID: "123",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(v.RequestID)
	// Output:
	// 123+
}
