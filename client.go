package klient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/worldline-go/logz"
	"golang.org/x/net/http2"
)

var (
	defaultMaxConnections = 100

	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4

	// DefaultBaseURL when empty "API_GATEWAY_ADDRESS" or "KLIENT_BASE_URL" env value will be used.
	DefaultBaseURL = ""

	// DisableEnvValues when true will disable all env values check.
	DisableEnvValues = false

	EnvKlientBaseURL            = "KLIENT_BASE_URL"
	EnvKlientBaseURLGlobal      = "API_GATEWAY_ADDRESS"
	EnvKlientInsecureSkipVerify = "KLIENT_INSECURE_SKIP_VERIFY"
	EnvKlientTimeout            = "KLIENT_TIMEOUT"
	EnvKlientRetryDisable       = "KLIENT_RETRY_DISABLE"
)

type Client struct {
	HTTP *http.Client
}

// NewPlain creates a new http client with the some default disabled automatic features.
//   - klient.WithDisableBaseURLCheck(true)
//   - klient.WithDisableRetry(true)
//   - klient.WithDisableEnvValues(true)
func NewPlain(opts ...OptionClientFn) (*Client, error) {
	opts = append([]OptionClientFn{
		WithDisableBaseURLCheck(true),
		WithDisableRetry(true),
		WithDisableEnvValues(true),
	}, opts...)

	return New(opts...)
}

// New creates a new http client with the provided options.
//
// Default BaseURL is required, it can be disabled by setting DisableBaseURLCheck to true.
func New(opts ...OptionClientFn) (*Client, error) {
	logAdapter := logz.AdapterKV{Log: log.Logger, Caller: true}
	o := optionClientValue{
		PooledClient:   true,
		MaxConnections: defaultMaxConnections,
		RetryWaitMin:   defaultRetryWaitMin,
		RetryWaitMax:   defaultRetryWaitMax,
		RetryMax:       defaultRetryMax,
		Backoff:        retryablehttp.DefaultBackoff,
		Logger:         logAdapter,
		RetryLog:       true,
	}

	for _, opt := range opts {
		opt(&o)
	}

	if o.RetryPolicy == nil {
		if o.RetryLog {
			options := []OptionRetryFn{
				OptionRetry.WithRetryLog(o.Logger),
			}
			options = append(options, o.OptionRetryFns...)

			o.RetryPolicy = NewRetryPolicy(options...)
		} else {
			o.RetryPolicy = NewRetryPolicy(o.OptionRetryFns...)
		}
	}

	if DisableEnvValues {
		o.DisableEnvValues = true
	}

	var baseURL *url.URL
	if o.BaseURL == "" {
		baseURL := DefaultBaseURL

		if !o.DisableEnvValues {
			if baseURL == "" {
				baseURL = os.Getenv(EnvKlientBaseURL)
			}
			if baseURL == "" {
				baseURL = os.Getenv(EnvKlientBaseURLGlobal)
			}
		}

		o.BaseURL = baseURL
	}

	if !o.DisableBaseURLCheck {
		if o.BaseURL == "" {
			return nil, fmt.Errorf("base url is required")
		}
	}

	if o.BaseURL != "" {
		var err error
		baseURL, err = url.Parse(o.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse base url: %w", err)
		}
	}

	// create client
	client := o.HTTPClient
	if client == nil {
		switch {
		case o.HTTP2:
			client = &http.Client{
				Transport: &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
						return net.Dial(network, addr)
					},
					IdleConnTimeout: 90 * time.Second,
				},
			}
		case o.PooledClient:
			client = cleanhttp.DefaultPooledClient()
		default:
			client = cleanhttp.DefaultClient()
		}
	}

	if o.BaseTransport != nil {
		client.Transport = o.BaseTransport
	}

	// make always after client creation
	if !o.HTTP2 && o.Proxy != "" {
		u, err := url.Parse(o.Proxy)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy url: %w", err)
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("failed to cast transport to http.Transport")
		}

		transport.Proxy = http.ProxyURL(u)
	}

	// skip verify
	if !o.DisableEnvValues {
		if v, _ := strconv.ParseBool(os.Getenv(EnvKlientInsecureSkipVerify)); v {
			o.InsecureSkipVerify = true
		}
	}

	if o.TLSConfig != nil {
		tlsClientConfig, err := o.TLSConfig.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate tls config: %w", err)
		}

		if o.HTTP2 {
			if transport, ok := client.Transport.(*http2.Transport); ok {
				transport.TLSClientConfig = tlsClientConfig
			}
		} else {
			if transport, ok := client.Transport.(*http.Transport); ok {
				transport.TLSClientConfig = tlsClientConfig
			}
		}
	}

	if o.InsecureSkipVerify {
		if o.HTTP2 {
			if transport, ok := client.Transport.(*http2.Transport); ok {
				tlsClientConfig := transport.TLSClientConfig
				if tlsClientConfig == nil {
					tlsClientConfig = &tls.Config{
						//nolint:gosec // user defined
						InsecureSkipVerify: true,
					}
				} else {
					tlsClientConfig.InsecureSkipVerify = true
				}

				transport.TLSClientConfig = tlsClientConfig
			}
		} else if transport, ok := client.Transport.(*http.Transport); ok {
			tlsClientConfig := transport.TLSClientConfig
			if tlsClientConfig == nil {
				tlsClientConfig = &tls.Config{
					//nolint:gosec // user defined
					InsecureSkipVerify: true,
				}
			} else {
				tlsClientConfig.InsecureSkipVerify = true
			}

			transport.TLSClientConfig = tlsClientConfig
		}
	}

	// disable
	if !o.DisableEnvValues {
		if v, _ := strconv.ParseBool(os.Getenv(EnvKlientRetryDisable)); v {
			o.DisableRetry = true
		}
	}

	if !o.DisableRetry {
		// create retry client
		retryClient := retryablehttp.Client{
			HTTPClient:   client,
			Logger:       o.Logger,
			RetryWaitMin: o.RetryWaitMin,
			RetryWaitMax: o.RetryWaitMax,
			RetryMax:     o.RetryMax,
			CheckRetry:   o.RetryPolicy,
			Backoff:      o.Backoff,
			ErrorHandler: PassthroughErrorHandler,
		}

		client = retryClient.StandardClient()
	}

	client.Transport = &TransportKlient{
		Base:    client.Transport,
		Header:  o.Header,
		BaseURL: baseURL,
		Inject:  o.Inject,
	}

	if len(o.RoundTripperList) > 0 {
		ctx := o.Ctx
		if ctx == nil {
			ctx = context.Background()
		}

		for _, roundTripper := range o.RoundTripperList {
			if roundTripper == nil {
				continue
			}

			transport, err := roundTripper(ctx, client.Transport)
			if err != nil {
				return nil, fmt.Errorf("failed to wrap transport: %w", err)
			}

			client.Transport = transport
		}
	}

	// set timeout
	if !o.DisableEnvValues {
		if v, _ := time.ParseDuration(os.Getenv(EnvKlientTimeout)); v > 0 {
			o.Timeout = v
		}
	}

	if o.Timeout > 0 {
		client.Timeout = o.Timeout
	}

	return &Client{
		HTTP: client,
	}, nil
}
