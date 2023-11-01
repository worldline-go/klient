package klient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/worldline-go/logz"
)

var (
	defaultMaxConnections = 100

	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4

	// DefaultBaseURL when empty os.Getenv("API_GATEWAY_ADDRESS") will use.
	DefaultBaseURL = ""
)

type Client struct {
	HTTP *http.Client
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
				OptionRetry.WithRetryLog(logAdapter),
			}
			options = append(options, o.OptionRetryFns...)

			o.RetryPolicy = NewRetryPolicy(options...)
		} else {
			o.RetryPolicy = NewRetryPolicy(o.OptionRetryFns...)
		}
	}

	var baseURL *url.URL
	if o.BaseURL == "" {
		baseURL := DefaultBaseURL
		if baseURL == "" {
			baseURL = os.Getenv("KLIENT_BASE_URL")
		}
		if baseURL == "" {
			baseURL = os.Getenv("API_GATEWAY_ADDRESS")
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
		if o.PooledClient {
			client = cleanhttp.DefaultPooledClient()
		} else {
			client = cleanhttp.DefaultClient()
		}
	}

	// disable retry
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
		}

		client = retryClient.StandardClient()
	}

	// skip verify
	if v, _ := strconv.ParseBool(os.Getenv("KLIENT_INSECURE_SKIP_VERIFY")); v {
		o.InsecureSkipVerify = true
	}

	if o.InsecureSkipVerify {
		//nolint:forcetypeassert // clear
		tlsClientConfig := client.Transport.(*http.Transport).TLSClientConfig
		if tlsClientConfig == nil {
			tlsClientConfig = &tls.Config{
				//nolint:gosec // user defined
				InsecureSkipVerify: true,
			}
		} else {
			tlsClientConfig.InsecureSkipVerify = true
		}

		//nolint:forcetypeassert // clear
		client.Transport.(*http.Transport).TLSClientConfig = tlsClientConfig
	}

	client.Transport = &TransportKlient{
		Base:    client.Transport,
		Header:  o.Header,
		BaseURL: baseURL,
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
	if v, _ := time.ParseDuration(os.Getenv("KLIENT_TIMEOUT")); v > 0 {
		o.Timeout = v
	}

	if o.Timeout > 0 {
		client.Timeout = o.Timeout
	}

	return &Client{
		HTTP: client,
	}, nil
}
