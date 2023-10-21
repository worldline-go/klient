package klient

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/worldline-go/logz"
)

type RoundTripperFunc = func(context.Context, http.RoundTripper) (http.RoundTripper, error)

type optionClientValue struct {
	// HTTPClient is the http client.
	HTTPClient *http.Client
	// PooledClient is generate pooled client if no http client provided.
	// Default is true.
	PooledClient bool
	// RoundTripperList is functions that wraps the default transport.
	//
	// Last function will be called first.
	RoundTripperList []RoundTripperFunc
	// Ctx for RoundTripper.
	Ctx context.Context
	// MaxConnections is the maximum number of idle (keep-alive) connections.
	MaxConnections int
	// Logger is the customer logger instance of retryablehttp. Can be either Logger or LeveledLogger
	Logger interface{}
	// InsecureSkipVerify is the flag to skip TLS verification.
	InsecureSkipVerify bool

	// Timeout is the timeout for the request.
	Timeout time.Duration

	// BaseURL is the base URL of the service.
	BaseURL string
	// DisableBaseURLCheck is the flag to disable base URL check.
	DisableBaseURLCheck bool

	// Header for default header to set if not exist.
	Header http.Header

	// DisableRetry is the flag to disable retry.
	DisableRetry bool
	// RetryWaitMin is the minimum wait time.
	// Default is 1 * time.Second.
	RetryWaitMin time.Duration
	// RetryWaitMax is the maximum wait time.
	// Default is 30 * time.Second.
	RetryWaitMax time.Duration
	// RetryMax is the maximum number of retry.
	// Default is 4.
	RetryMax int
	// RetryPolicy is the retry policy.
	RetryPolicy retryablehttp.CheckRetry
	// Backoff is the backoff policy.
	Backoff retryablehttp.Backoff
	// RetryLog is the flag to enable retry log of the http body. Default is true.
	RetryLog bool
	// OptionRetryFns is the retry options for default retry policy.
	OptionRetryFns []OptionRetryFn
}

// OptionClientFn is a function that configures the client.
type OptionClientFn func(*optionClientValue)

// WithHTTPClient configures the client to use the provided http client.
func WithHTTPClient(httpClient *http.Client) OptionClientFn {
	return func(o *optionClientValue) {
		o.HTTPClient = httpClient
	}
}

func WithPooledClient(pooledClient bool) OptionClientFn {
	return func(o *optionClientValue) {
		o.PooledClient = pooledClient
	}
}

// WithRoundTripper configures the client to wrap the default transport.
func WithRoundTripper(f ...func(context.Context, http.RoundTripper) (http.RoundTripper, error)) OptionClientFn {
	return func(o *optionClientValue) {
		o.RoundTripperList = append(o.RoundTripperList, f...)
	}
}

// WithCtx for TransportWrapper call.
func WithCtx(ctx context.Context) OptionClientFn {
	return func(o *optionClientValue) {
		o.Ctx = ctx
	}
}

// WithHeader configures the client to use this default header if not exist.
func WithHeader(header http.Header) OptionClientFn {
	return func(o *optionClientValue) {
		o.Header = header
	}
}

// WithMaxConnections configures the client to use the provided maximum number of idle connections.
func WithMaxConnections(maxConnections int) OptionClientFn {
	return func(o *optionClientValue) {
		o.MaxConnections = maxConnections
	}
}

// WithLogger configures the client to use the provided logger.
//
// For zerolog logz.AdapterKV{Log: logger} can usable.
func WithLogger(logger logz.Adapter) OptionClientFn {
	return func(o *optionClientValue) {
		o.Logger = logger
	}
}

// WithInsecureSkipVerify configures the client to skip TLS verification.
func WithInsecureSkipVerify(insecureSkipVerify bool) OptionClientFn {
	return func(o *optionClientValue) {
		o.InsecureSkipVerify = insecureSkipVerify
	}
}

// WithBaseURL configures the client to use the provided base URL.
func WithBaseURL(baseURL string) OptionClientFn {
	return func(o *optionClientValue) {
		o.BaseURL = baseURL
	}
}

// WithDisableBaseURLCheck configures the client to disable base URL check.
func WithDisableBaseURLCheck(baseURLCheck bool) OptionClientFn {
	return func(o *optionClientValue) {
		o.DisableBaseURLCheck = baseURLCheck
	}
}

// WithTimeout configures the client to use the provided timeout. Default is no timeout.
//
// Warning: This timeout is for the whole request, including retries.
func WithTimeout(timeout time.Duration) OptionClientFn {
	return func(o *optionClientValue) {
		o.Timeout = timeout
	}
}

// WithDisableRetry configures the client to disable retry.
func WithDisableRetry(disableRetry bool) OptionClientFn {
	return func(options *optionClientValue) {
		options.DisableRetry = disableRetry
	}
}

// WithRetryWaitMin configures the client to use the provided minimum wait time.
func WithRetryWaitMin(retryWaitMin time.Duration) OptionClientFn {
	return func(options *optionClientValue) {
		options.RetryWaitMin = retryWaitMin
	}
}

// WithRetryWaitMax configures the client to use the provided maximum wait time.
func WithRetryWaitMax(retryWaitMax time.Duration) OptionClientFn {
	return func(options *optionClientValue) {
		options.RetryWaitMax = retryWaitMax
	}
}

// WithRetryMax configures the client to use the provided maximum number of retry.
func WithRetryMax(retryMax int) OptionClientFn {
	return func(options *optionClientValue) {
		options.RetryMax = retryMax
	}
}

// WithBackoff configures the client to use the provided backoff.
func WithBackoff(backoff retryablehttp.Backoff) OptionClientFn {
	return func(options *optionClientValue) {
		options.Backoff = backoff
	}
}

// WithRetryPolicy configures the client to use the provided retry policy.
func WithRetryPolicy(retryPolicy retryablehttp.CheckRetry) OptionClientFn {
	return func(options *optionClientValue) {
		options.RetryPolicy = retryPolicy
	}
}

// WithRetryLog configures the client to use the provided retry log flag, default is true.
//
// This option is only used with default retry policy.
func WithRetryLog(retryLog bool) OptionClientFn {
	return func(options *optionClientValue) {
		options.RetryLog = retryLog
	}
}

func WithRetryOptions(opts ...OptionRetryFn) OptionClientFn {
	return func(options *optionClientValue) {
		options.OptionRetryFns = append(options.OptionRetryFns, opts...)
	}
}
