package klient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/rs/zerolog/log"
)

type TestRequest struct {
	ID string `json:"id"`

	header http.Header `json:"-"`
}

func (TestRequest) Method() string {
	return http.MethodPost
}

func (TestRequest) Path() string {
	return "/api/v1/test"
}

func (r TestRequest) BodyJSON() interface{} {
	return r
}

func (r TestRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("id is required")
	}

	return nil
}

func (r TestRequest) Header() http.Header {
	header := http.Header{}
	header.Set("X-Info", "test")

	for k, v := range r.header {
		header[k] = v
	}

	return header
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
		req  InfRequest
		resp *map[string]interface{}
	}
	tests := []struct {
		name        string
		args        args
		want        interface{}
		wantErr     bool
		retryCount  int
		short       bool
		long        bool
		optionRetry []optionRetryFn
		extraCheck  func(r *http.Request) error
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
			name: "DoWithFunc with retry disable",
			args: args{
				ctx: context.Background(),
				req: TestRequest{
					ID: "123",
				},
				resp: new(map[string]interface{}),
			},
			optionRetry: []optionRetryFn{
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
			optionRetry: []optionRetryFn{
				// OptionRetry.WithRetryDisable(),
			},
			want: map[string]interface{}{
				"request_id": "123+",
			},
			wantErr:    false,
			retryCount: 2,
			long:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.long && testing.Short() {
				t.Skip("skipping test in short mode.")
			}

			retryCount = tt.retryCount
			extraCheck = tt.extraCheck

			httpxClient, err := NewClient(
				OptionClient.WithBaseURL(httpServer.URL),
				OptionClient.WithRetryMax(tt.retryCount),
				OptionClient.WithRetryOptions(tt.optionRetry...),
			)
			if err != nil {
				t.Errorf("NewClient() error = %v", err)
				return
			}

			err = httpxClient.DoWithFunc(context.Background(), tt.args.req, func(r *http.Response) error {
				if err := UnexpectedResponse(r); err != nil {
					return err
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					return err
				}

				log.Info().Msgf("response: %s", body)

				if err := json.Unmarshal(body, &tt.args.resp); err != nil {
					return err
				}

				return nil
			})
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
