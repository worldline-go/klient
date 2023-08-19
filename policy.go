package klient

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

var ResponseErrLimit int64 = 1 << 20 // 1MB

type optionRetry struct {
	DisableRetry        Null[bool]
	DisabledStatusCodes []int
	EnabledStatusCodes  []int
}

type OptionRetryValue = optionRetry

type OptionRetryFn func(*optionRetry)

type OptionRetry struct{}

// WithOptionRetry configures the retry policy directly.
//
// This option overrides all other retry options when previously set.
func (OptionRetry) WithOptionRetry(oRetry *OptionRetryValue) OptionRetryFn {
	return func(o *optionRetry) {
		*o = *oRetry
	}
}

func (OptionRetry) WithRetryDisable() OptionRetryFn {
	return func(o *optionRetry) {
		o.DisableRetry = Null[bool]{Value: true, Valid: true}
	}
}

func (OptionRetry) WithRetryDisabledStatusCodes(codes ...int) OptionRetryFn {
	return func(o *optionRetry) {
		o.DisabledStatusCodes = append(o.DisabledStatusCodes, codes...)
	}
}

func (OptionRetry) WithRetryEnabledStatusCodes(codes ...int) OptionRetryFn {
	return func(o *optionRetry) {
		o.EnabledStatusCodes = append(o.EnabledStatusCodes, codes...)
	}
}

type ctxValueType string

const (
	optionsRetryCtxValue ctxValueType = "retry"
)

// RequestCtx adds the request options to the context.
func RetryPolicyCtx(ctx context.Context, opts ...OptionRetryFn) context.Context {
	o := optionRetry{}

	for _, opt := range opts {
		opt(&o)
	}

	return context.WithValue(ctx, optionsRetryCtxValue, o)
}

// RetryPolicy provides a default callback for Client.CheckRetry, which
// will retry on connection errors and server errors.
func RetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	retryValue := ctx.Value(optionsRetryCtxValue)
	if retryValue != nil {
		retryValue, ok := retryValue.(optionRetry)
		if ok {
			if retryValue.DisableRetry.Valid && retryValue.DisableRetry.Value {
				return false, nil
			}

			for _, disabledStatusCode := range retryValue.DisabledStatusCodes {
				if resp.StatusCode == disabledStatusCode {
					return false, nil
				}
			}

			for _, enabledStatusCode := range retryValue.EnabledStatusCodes {
				if resp.StatusCode == enabledStatusCode {
					return true, fmt.Errorf("force retried HTTP status %s: [%s]", resp.Status, LimitedResponse(resp))
				}
			}
		}
	}

	v, errPolicy := retryablehttp.ErrorPropagatedRetryPolicy(ctx, resp, err)
	if v && errPolicy != nil {
		err = fmt.Errorf("%w: [%s]; previous error: %w", err, LimitedResponse(resp), err)
	}

	return v, err
}

// LimitedResponse not close body, retry library draining it.
func LimitedResponse(resp *http.Response) []byte {
	v, _ := io.ReadAll(io.LimitReader(resp.Body, ResponseErrLimit))

	return v
}
