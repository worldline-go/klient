package klient

import "time"

type Config struct {
	BaseURL string `cfg:"base_url"`

	Header map[string][]string `cfg:"header"`

	Timeout             time.Duration `cfg:"timeout"`
	DisableBaseURLCheck *bool         `cfg:"disable_base_url_check"`
	DisableEnvValues    *bool         `cfg:"disable_env_values"`
	InsecureSkipVerify  *bool         `cfg:"insecure_skip_verify"`

	DisableRetry *bool         `cfg:"disable_retry"`
	RetryMax     int           `cfg:"retry_max"`
	RetryWaitMin time.Duration `cfg:"retry_wait_min"`
	RetryWaitMax time.Duration `cfg:"retry_wait_max"`

	Proxy string `cfg:"proxy"`
	HTTP2 *bool  `cfg:"http2"`

	TLSConfig *TLSConfig `cfg:"tls"`
}

func (c Config) Options(opts ...OptionClientFn) []OptionClientFn {
	if c.BaseURL != "" {
		opts = append(opts, WithBaseURL(c.BaseURL))
	}

	if c.Timeout != 0 {
		opts = append(opts, WithTimeout(c.Timeout))
	}

	if c.DisableBaseURLCheck != nil {
		opts = append(opts, WithDisableBaseURLCheck(*c.DisableBaseURLCheck))
	}

	if c.DisableEnvValues != nil {
		opts = append(opts, WithDisableEnvValues(*c.DisableEnvValues))
	}

	if c.InsecureSkipVerify != nil {
		opts = append(opts, WithInsecureSkipVerify(*c.InsecureSkipVerify))
	}

	if c.DisableRetry != nil {
		opts = append(opts, WithDisableRetry(*c.DisableRetry))
	}

	if c.RetryMax != 0 {
		opts = append(opts, WithRetryMax(c.RetryMax))
	}

	if c.RetryWaitMin != 0 {
		opts = append(opts, WithRetryWaitMin(c.RetryWaitMin))
	}

	if c.RetryWaitMax != 0 {
		opts = append(opts, WithRetryWaitMax(c.RetryWaitMax))
	}

	if len(c.Header) > 0 {
		opts = append(opts, WithHeaderSet(c.Header))
	}

	if c.Proxy != "" {
		opts = append(opts, WithProxy(c.Proxy))
	}

	if c.HTTP2 != nil {
		opts = append(opts, WithHTTP2(*c.HTTP2))
	}

	if c.TLSConfig != nil {
		opts = append(opts, WithTLSConfig(c.TLSConfig))
	}

	return opts
}

func (c Config) New(options ...OptionClientFn) (*Client, error) {
	return New(c.Options(options...)...)
}
