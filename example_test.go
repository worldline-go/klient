package klient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

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

	// os.Setenv("API_GATEWAY_ADDRESS", httpServer.URL)
	klient.DefaultBaseURL = httpServer.URL

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

func ExampleWithRetryTimeout() {
	// Create a client with retry timeout of 2 seconds per attempt
	client, err := klient.New(
		klient.WithDisableBaseURLCheck(true),
		klient.WithRetryMax(3), // Will retry up to 3 times
		klient.WithRetryWaitMin(500*time.Millisecond),
		klient.WithRetryWaitMax(1*time.Second),
		klient.WithRetryTimeout(2*time.Second), // Each attempt times out after 2 seconds
	)
	if err != nil {
		panic(err)
	}

	// If a request takes longer than 2 seconds, it will timeout and retry
	// Total possible time: ~2s (first attempt) + 0.5s (wait) + 2s (retry) + ...
	_ = client
	fmt.Println("Client created with 2 second timeout per retry attempt")
	// Output: Client created with 2 second timeout per retry attempt
}

func TestRetryTimeout(t *testing.T) {
	var attemptCount atomic.Int32

	// Create a server that delays responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attemptCount.Add(1)
		t.Logf("Attempt %d received", attempt)

		// First 2 attempts take too long (will timeout)
		// Third attempt is fast (will succeed)
		if attempt < 3 {
			time.Sleep(3 * time.Second) // Longer than RetryTimeout
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Create client with retry timeout
	client, err := klient.New(
		klient.WithDisableBaseURLCheck(true),
		klient.WithRetryMax(3),
		klient.WithRetryWaitMin(100*time.Millisecond),
		klient.WithRetryWaitMax(200*time.Millisecond),
		klient.WithRetryTimeout(2*time.Second), // Each attempt times out after 2 seconds
	)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	resp, err := client.HTTP.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Request failed after %v: %v", elapsed, err)
	}
	defer resp.Body.Close()

	attempts := attemptCount.Load()
	t.Logf("Request completed in %v with %d attempts", elapsed, attempts)

	// Should have made 3 attempts (2 timeouts + 1 success)
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Total time should be approximately:
	// 2s (timeout) + 0.1-0.2s (wait) + 2s (timeout) + 0.1-0.2s (wait) + <1s (success)
	// = approximately 4-5 seconds
	if elapsed < 4*time.Second || elapsed > 6*time.Second {
		t.Logf("Warning: elapsed time %v outside expected range (4-6s)", elapsed)
	}
}

func TestRetryTimeoutAllAttemptsTimeout(t *testing.T) {
	var attemptCount atomic.Int32

	// Create a server that always delays
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attemptCount.Add(1)
		t.Logf("Attempt %d received", attempt)
		time.Sleep(5 * time.Second) // Always too slow
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := klient.New(
		klient.WithDisableBaseURLCheck(true),
		klient.WithRetryMax(2), // Will try 3 times total (initial + 2 retries)
		klient.WithRetryWaitMin(100*time.Millisecond),
		klient.WithRetryWaitMax(200*time.Millisecond),
		klient.WithRetryTimeout(1*time.Second), // Each attempt times out after 1 second
	)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	resp, err := client.HTTP.Do(req)
	elapsed := time.Since(start)

	if resp != nil {
		resp.Body.Close()
	}

	attempts := attemptCount.Load()
	t.Logf("Request failed after %v with %d attempts (error: %v)", elapsed, attempts, err)

	// Should have made 3 attempts (all timeouts)
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Should have an error (context deadline exceeded)
	if err == nil {
		t.Error("Expected error due to timeouts, got nil")
	}

	// Total time should be approximately:
	// 1s (timeout) + 0.1-0.2s (wait) + 1s (timeout) + 0.1-0.2s (wait) + 1s (timeout)
	// = approximately 3-4 seconds
	if elapsed < 3*time.Second || elapsed > 4*time.Second {
		t.Logf("Warning: elapsed time %v outside expected range (3-4s)", elapsed)
	}
}
