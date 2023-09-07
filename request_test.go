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
	"time"

	"github.com/go-test/deep"
	"github.com/rs/zerolog/log"
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

	client, err := New(
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
				HttpClient: client.HTTP,
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
				HTTP: tt.fields.HttpClient,
			}

			req, err := http.NewRequestWithContext(tt.args.ctx, tt.args.method, tt.args.path, tt.args.body)
			if err != nil {
				t.Fatalf("Client.Request() error = %v", err)
			}

			req.Header = tt.args.header

			if err := c.Do(req, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("Client.Request() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type TestRequest struct {
	ID string `json:"id"`

	header http.Header `json:"-"`
}

func (r TestRequest) Request(ctx context.Context) (*http.Request, error) {
	bodyData, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request body: %w", err)
	}

	body := bytes.NewReader(bodyData)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/test", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Info", "test")
	req.Header.Set("Content-Type", "application/json")

	for k, v := range r.header {
		req.Header[k] = v
	}

	return req, nil
}

func (TestRequest) Response(r *http.Response) (map[string]interface{}, error) {
	if err := UnexpectedResponse(r); err != nil {
		return nil, err
	}

	var m map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}

func TestClient_Do(t *testing.T) {
	retryCount := 0
	var extraCheck func(r *http.Request) error

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retry
		if retryCount > 0 {
			retryCount--
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "internal server error Ipsum eiusmod non officia cillum est ullamco qui est pariatur id. Tempor incididunt excepteur officia aute ullamco in incididunt. Dolor veniam reprehenderit non aliqua. Anim laboris ut commodo fugiat exercitation cupidatat exercitation consectetur aliqua consequat sint est eu occaecat. Esse qui exercitation magna consectetur pariatur ut adipisicing aute qui ad ea incididunt sint eu. Mollit sunt do ipsum sunt ex proident duis. Cupidatat cillum nulla sint cupidatat cupidatat enim et commodo duis qui sunt eiusmod commodo. Aliqua elit cupidatat nulla enim excepteur cupidatat tempor aliquip tempor consequat qui. Commodo veniam excepteur cillum Lorem minim. Magna ipsum veniam ipsum cillum. Mollit ullamco veniam qui elit quis duis amet laboris in eiusmod. Irure est adipisicing reprehenderit laboris occaecat anim. Excepteur fugiat laborum fugiat fugiat deserunt ut ex adipisicing culpa occaecat pariatur et aliqua. Duis proident officia sint adipisicing aute aute incididunt quis esse. Quis ex sint magna pariatur exercitation aliqua do reprehenderit occaecat aute est dolor voluptate reprehenderit. "}`))

			return
		}

		// check request method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid request method"}`))
			return
		}

		// check request path
		if r.URL.Path != "/api/v1/test" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid request path"}`))
			return
		}

		// check request header
		if r.Header.Get("X-Info") != "test" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid request header"}`))
			return
		}

		// get request body
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid request body"}`))
			return
		}

		if m["id"] == "444" {
			time.Sleep(500 * time.Millisecond)
			_, _ = w.Write([]byte(`{"error": "time sleep"}`))
			return
		}

		// check request body
		if m["id"] != "123" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid id"}`))
			return
		}

		// extra check
		if extraCheck != nil {
			if err := extraCheck(r); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))
				return
			}
		}

		// write response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"request_id": "123+"}`))
	}))

	defer httpServer.Close()

	type args struct {
		ctx  context.Context
		req  Requester[map[string]interface{}]
		resp *map[string]interface{}
	}
	tests := []struct {
		name         string
		args         args
		want         interface{}
		wantErr      bool
		retryCount   int
		short        bool
		long         bool
		optionRetry  []OptionRetryFn
		optionClient []OptionClientFn
		extraCheck   func(r *http.Request) error
	}{
		{
			name: "DoWithFunc",
			args: args{
				ctx: context.Background(),
				req: TestRequest{
					ID: "123",

					header: http.Header{
						"X-Ctx": []string{"test"},
					},
				},
				resp: new(map[string]interface{}),
			},
			extraCheck: func(r *http.Request) error {
				if v := r.Header.Get("X-Ctx"); v != "test" {
					return fmt.Errorf("invalid request header %s", v)
				}

				return nil
			},
			want: map[string]interface{}{
				"request_id": "123+",
			},
			wantErr: false,
		},
		{
			name: "DoWithFunc ctx header",
			args: args{
				ctx: context.WithValue(context.Background(), TransportHeaderKey, http.Header{
					"X-Ctx": []string{"test"},
				}),
				req: TestRequest{
					ID: "123",
				},
				resp: new(map[string]interface{}),
			},
			extraCheck: func(r *http.Request) error {
				if v := r.Header.Get("X-Ctx"); v != "test" {
					return fmt.Errorf("invalid request header %s", v)
				}

				return nil
			},
			want: map[string]interface{}{
				"request_id": "123+",
			},
			wantErr: false,
		},
		{
			name: "DoWithFunc with retry disable",
			args: args{
				ctx: context.Background(),
				req: TestRequest{
					ID: "123",
				},
				resp: new(map[string]interface{}),
			},
			optionRetry: []OptionRetryFn{
				OptionRetry.WithRetryDisable(),
			},
			wantErr:    true,
			retryCount: 5,
			// long:       true,
		},
		{
			name: "DoWithFunc with retry",
			args: args{
				ctx: context.Background(),
				req: TestRequest{
					ID: "123",
				},
				resp: new(map[string]interface{}),
			},
			optionRetry: []OptionRetryFn{
				// OptionRetry.WithRetryDisable(),
			},
			want: map[string]interface{}{
				"request_id": "123+",
			},
			wantErr:    false,
			retryCount: 2,
			long:       true,
		},
		{
			name: "Timeout test",
			args: args{
				ctx: context.Background(),
				req: TestRequest{
					ID: "444",
				},
				resp: new(map[string]interface{}),
			},
			optionClient: []OptionClientFn{
				OptionClient.WithDisableRetry(true),
				OptionClient.WithTimeout(100 * time.Millisecond),
			},
			want: map[string]interface{}{
				"request_id": "123+",
			},
			wantErr: true,
			long:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.long && testing.Short() {
				t.Skip("skipping test in short mode.")
			}

			retryCount = tt.retryCount
			extraCheck = tt.extraCheck

			tt.optionClient = append(
				tt.optionClient,
				OptionClient.WithBaseURL(httpServer.URL),
				OptionClient.WithRetryMax(tt.retryCount),
				OptionClient.WithRetryOptions(tt.optionRetry...),
			)

			client, err := New(tt.optionClient...)
			if err != nil {
				t.Errorf("NewClient() error = %v", err)
				return
			}

			ctx := tt.args.ctx
			if ctx == nil {
				ctx = context.Background()
			}

			*tt.args.resp, err = DoWithInf(ctx, client.HTTP, tt.args.req)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Client.DoWithFunc() error = %v, wantErr %v", err, tt.wantErr)
				}

				log.Error().Err(err).Msg("error")
			}

			if tt.wantErr && err == nil {
				t.Fatalf("Client.DoWithFunc() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				log.Info().Msgf("error: %v", err)

				return
			}

			// tt.args.resp is a pointer
			if diff := deep.Equal(*tt.args.resp, tt.want); diff != nil {
				t.Fatalf("Client.DoWithFunc() resp diff = %v", diff)
			}
		})
	}
}
