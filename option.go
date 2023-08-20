package klient

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"github.com/worldline-go/logz"
)

type optionClientValue struct {
	// HTTPClient is the http client.
	HTTPClient *http.Client
	// PooledClient is generate pooled client if no http client provided.
	// Default is true.
	PooledClient bool
	// TransportWrapper is a function that wraps the default transport.
	TransportWrapper func(context.Context, http.RoundTripper) (http.RoundTripper, error)
	// Ctx for TransportWrapper.
	Ctx context.Context
	// MaxConnections is the maximum number of idle (keep-alive) connections.
	MaxConnections int
	// Logger is the customer logger instance of retryablehttp. Can be either Logger or LeveledLogger
	Logger interface{}
	// InsecureSkipVerify is the flag to skip TLS verification.
	InsecureSkipVerify bool

	// BaseURL is the base URL of the service.
	BaseURL string
	// DisableBaseURLCheck is the flag to disable base URL check.
	DisableBaseURLCheck bool

	// DisableRetry is the flag to disable retry.
	DisableRetry bool
	// RetryWaitMin is the minimum wait time.
	// Default is 100ms.
	RetryWaitMin time.Duration
	// RetryWaitMax is the maximum wait time.
	RetryWaitMax time.Duration
	// RetryMax is the maximum number of retry.
	RetryMax int
	// RetryPolicy is the retry policy.
	RetryPolicy retryablehttp.CheckRetry
	// Backoff is the backoff policy.
	Backoff retryablehttp.Backoff
	// RetryLog is the flag to enable retry log of the http body. Default is true.
	RetryLog bool
	// OptionRetryFns is the retry options for default retry policy.
	OptionRetryFns []optionRetryFn
}

// optionClientFn is a function that configures the client.
type optionClientFn func(*optionClientValue)

type optionClient struct{}

var OptionClient = optionClient{}

// WithHTTPClient configures the client to use the provided http client.
func (optionClient) WithHTTPClient(httpClient *http.Client) optionClientFn {
	return func(o *optionClientValue) {
		o.HTTPClient = httpClient
	}
}

func (optionClient) WithPooledClient(pooledClient bool) optionClientFn {
	return func(o *optionClientValue) {
		o.PooledClient = pooledClient
	}
}

// WithTransportWrapper configures the client to wrap the default transport.
func (optionClient) WithTransportWrapper(f func(context.Context, http.RoundTripper) (http.RoundTripper, error)) optionClientFn {
	return func(o *optionClientValue) {
		o.TransportWrapper = f
	}
}

// WithCtx for TransportWrapper call.
func (optionClient) WithCtx(ctx context.Context) optionClientFn {
	return func(o *optionClientValue) {
		o.Ctx = ctx
	}
}

// WithMaxConnections configures the client to use the provided maximum number of idle connections.
func (optionClient) WithMaxConnections(maxConnections int) optionClientFn {
	return func(o *optionClientValue) {
		o.MaxConnections = maxConnections
	}
}

// WithLogger configures the client to use the provided logger.
func (optionClient) WithLogger(logger zerolog.Logger) optionClientFn {
	return func(o *optionClientValue) {
		o.Logger = logz.AdapterKV{Log: logger}
	}
}

// WithInsecureSkipVerify configures the client to skip TLS verification.
func (optionClient) WithInsecureSkipVerify(insecureSkipVerify bool) optionClientFn {
	return func(o *optionClientValue) {
		o.InsecureSkipVerify = insecureSkipVerify
	}
}

// WithBaseURL configures the client to use the provided base URL.
func (optionClient) WithBaseURL(baseURL string) optionClientFn {
	return func(o *optionClientValue) {
		o.BaseURL = baseURL
	}
}

// WithDisableBaseURLCheck configures the client to disable base URL check.
func (optionClient) WithDisableBaseURLCheck(baseURLCheck bool) optionClientFn {
	return func(o *optionClientValue) {
		o.DisableBaseURLCheck = baseURLCheck
	}
}

// WithDisableRetry configures the client to disable retry.
func (optionClient) WithDisableRetry(disableRetry bool) optionClientFn {
	return func(options *optionClientValue) {
		options.DisableRetry = disableRetry
	}
}

// WithRetryWaitMin configures the client to use the provided minimum wait time.
func (optionClient) WithRetryWaitMin(retryWaitMin time.Duration) optionClientFn {
	return func(options *optionClientValue) {
		options.RetryWaitMin = retryWaitMin
	}
}

// WithRetryWaitMax configures the client to use the provided maximum wait time.
func (optionClient) WithRetryWaitMax(retryWaitMax time.Duration) optionClientFn {
	return func(options *optionClientValue) {
		options.RetryWaitMax = retryWaitMax
	}
}

// WithRetryMax configures the client to use the provided maximum number of retry.
func (optionClient) WithRetryMax(retryMax int) optionClientFn {
	return func(options *optionClientValue) {
		options.RetryMax = retryMax
	}
}

// WithBackoff configures the client to use the provided backoff.
func (optionClient) WithBackoff(backoff retryablehttp.Backoff) optionClientFn {
	return func(options *optionClientValue) {
		options.Backoff = backoff
	}
}

// WithRetryPolicy configures the client to use the provided retry policy.
func (optionClient) WithRetryPolicy(retryPolicy retryablehttp.CheckRetry) optionClientFn {
	return func(options *optionClientValue) {
		options.RetryPolicy = retryPolicy
	}
}

// WithRetryLog configures the client to use the provided retry log flag, default is true.
//
// This option is only used with default retry policy.
func (optionClient) WithRetryLog(retryLog bool) optionClientFn {
	return func(options *optionClientValue) {
		options.RetryLog = retryLog
	}
}

func (optionClient) WithRetryOptions(opts ...optionRetryFn) optionClientFn {
	return func(options *optionClientValue) {
		options.OptionRetryFns = append(options.OptionRetryFns, opts...)
	}
}
