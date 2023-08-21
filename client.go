package klient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
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
)

type Client struct {
	HTTPClient      *http.Client
	retryableClient *retryablehttp.Client
	BaseURL         *url.URL
}

// New creates a new http client with the provided options.
//
// Default BaseURL is required, it can be disabled by setting DisableBaseURLCheck to true.
func New(opts ...OptionClientFn) (*Client, error) {
	logAdapter := logz.AdapterKV{Log: log.Logger}
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
	if !o.DisableBaseURLCheck {
		if o.BaseURL == "" {
			return nil, fmt.Errorf("base url is required")
		}

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

	if o.TransportWrapper != nil {
		ctx := o.Ctx
		if ctx == nil {
			ctx = context.Background()
		}

		transport, err := o.TransportWrapper(ctx, client.Transport)
		if err != nil {
			return nil, fmt.Errorf("failed to wrap transport: %w", err)
		}

		client.Transport = transport
	}

	var retryClient *retryablehttp.Client
	// disable retry
	if !o.DisableRetry {
		// create retry client
		retryClient = &retryablehttp.Client{
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

	return &Client{
		HTTPClient:      client,
		retryableClient: retryClient,
		BaseURL:         baseURL,
	}, nil
}
