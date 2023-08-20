package klient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/worldline-go/klient"
)

type Client struct {
	klient *klient.Client
}

func (c *Client) CreateX(ctx context.Context, r CreateXRequest) (*CreateXResponse, error) {
	var v CreateXResponse
	if err := c.klient.DoWithFunc(ctx, r, func(r *http.Response) error {
		if r.StatusCode != http.StatusOK {
			return klient.UnexpectedResponseError(r)
		}

		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &v, nil
}

// ----

type CreateXRequest struct {
	ID string `json:"id"`
}

func (CreateXRequest) Method() string {
	return http.MethodPost
}

func (CreateXRequest) Path() string {
	return "/api/v1/x"
}

func (r CreateXRequest) BodyJSON() interface{} {
	return r
}

func (r CreateXRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id is required")
	}

	return nil
}

func (r CreateXRequest) Header() http.Header {
	v := http.Header{}
	v.Set("X-Info", "example")

	return v
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

	httpxClient, err := klient.New(
		klient.OptionClient.WithBaseURL(httpServer.URL),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := &Client{
		klient: httpxClient,
	}

	v, err := c.CreateX(context.Background(), CreateXRequest{
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
