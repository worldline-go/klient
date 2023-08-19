package klient

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
)

var (
	defaultMaxConnections = 100
	// Default retry configuration
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
}

// NewClient creates a new http client with the provided options.
//
// Default BaseURL is required, it can be disabled by setting DisableBaseURLCheck to true.
func NewClient(opts ...optionClientFn) (*Client, error) {
	o := optionClientValue{
		PooledClient:   true,
		MaxConnections: defaultMaxConnections,
		RetryWaitMin:   defaultRetryWaitMin,
		RetryWaitMax:   defaultRetryWaitMax,
		RetryMax:       defaultRetryMax,
		RetryPolicy:    RetryPolicy,
		Backoff:        retryablehttp.DefaultBackoff,
		Logger:         log.New(os.Stderr, "", log.LstdFlags),
	}

	for _, opt := range opts {
		opt(&o)
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
		transport, err := o.TransportWrapper(o.Ctx, client.Transport)
		if err != nil {
			return nil, fmt.Errorf("failed to wrap transport: %w", err)
		}

		client.Transport = transport
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

	return &Client{
		HTTPClient: client,
		BaseURL:    baseURL,
	}, nil
}
