package klient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/worldline-go/logz"
)

type Null[T any] struct {
	Value T
	Valid bool
}

type optionRetryValue struct {
	DisableRetry        Null[bool]
	DisabledStatusCodes []int
	EnabledStatusCodes  []int
	Log                 logz.Adapter
}

type OptionRetryValue = optionRetryValue

type OptionRetryFn func(*optionRetryValue)

type OptionRetryHolder struct{}

var OptionRetry = OptionRetryHolder{}

// WithOptionRetry configures the retry policy directly.
//
// This option overrides all other retry options when previously set.
func (OptionRetryHolder) WithOptionRetry(oRetry *OptionRetryValue) OptionRetryFn {
	return func(o *optionRetryValue) {
		*o = *oRetry
	}
}

func (OptionRetryHolder) WithRetryDisable() OptionRetryFn {
	return func(o *optionRetryValue) {
		o.DisableRetry = Null[bool]{Value: true, Valid: true}
	}
}

func (OptionRetryHolder) WithRetryDisabledStatusCodes(codes ...int) OptionRetryFn {
	return func(o *optionRetryValue) {
		o.DisabledStatusCodes = append(o.DisabledStatusCodes, codes...)
	}
}

func (OptionRetryHolder) WithRetryEnabledStatusCodes(codes ...int) OptionRetryFn {
	return func(o *optionRetryValue) {
		o.EnabledStatusCodes = append(o.EnabledStatusCodes, codes...)
	}
}

func (OptionRetryHolder) WithRetryLog(log logz.Adapter) OptionRetryFn {
	return func(o *optionRetryValue) {
		o.Log = log
	}
}

func NewRetryPolicy(opts ...OptionRetryFn) retryablehttp.CheckRetry {
	o := optionRetryValue{}

	for _, opt := range opts {
		opt(&o)
	}

	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		return retryPolicyOpts(ctx, resp, err, &o)
	}
}

// RetryPolicy provides a default callback for Client.CheckRetry, which
// will retry on connection errors and server errors.
func RetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	return retryPolicyOpts(ctx, resp, err, nil)
}

// RetryPolicy provides a default callback for Client.CheckRetry, which
// will retry on connection errors and server errors.
func retryPolicyOpts(ctx context.Context, resp *http.Response, err error, retryValue *optionRetryValue) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if err := ctx.Err(); err != nil {
		return false, err
	}

	if retryValue != nil {
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

	v, errPolicy := retryablehttp.ErrorPropagatedRetryPolicy(ctx, resp, err)

	return retryError(v, errPolicy, resp, retryValue.Log, err)
}

func retryError(retry bool, err error, resp *http.Response, log logz.Adapter, errOrg error) (bool, error) {
	if !retry {
		return retry, err
	}

	response := LimitedResponse(resp)
	if log != nil {
		errLog := err
		if errLog == nil {
			errLog = errOrg
		}
		log.Warn("retrying request", "response", string(response), "error", errLog)
	}

	if err == nil {
		return retry, nil
	}

	return retry, fmt.Errorf("%w: [%s]", err, response)
}

func PassthroughErrorHandler(resp *http.Response, err error, _ int) (*http.Response, error) {
	if resp == nil {
		return nil, err
	}

	return resp, nil
}
