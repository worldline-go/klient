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
	RetryTimeout time.Duration `cfg:"retry_timeout"`

	PooledClient *bool `cfg:"pooled_client"`

	Proxy string `cfg:"proxy"`
	HTTP2 *bool  `cfg:"http2"`

	TLSConfig *TLSConfig `cfg:"tls"`
}

func (c Config) ToOption() OptionClientFn {
	return func(o *optionClientValue) {
		if c.BaseURL != "" {
			o.BaseURL = c.BaseURL
		}

		if c.Timeout != 0 {
			o.Timeout = c.Timeout
		}

		if c.DisableBaseURLCheck != nil {
			o.DisableBaseURLCheck = *c.DisableBaseURLCheck
		}

		if c.DisableEnvValues != nil {
			o.DisableEnvValues = *c.DisableEnvValues
		}

		if c.InsecureSkipVerify != nil {
			o.InsecureSkipVerify = *c.InsecureSkipVerify
		}

		if c.DisableRetry != nil {
			o.DisableRetry = *c.DisableRetry
		}

		if c.RetryMax != 0 {
			o.RetryMax = c.RetryMax
		}

		if c.RetryWaitMin != 0 {
			o.RetryWaitMin = c.RetryWaitMin
		}

		if c.RetryWaitMax != 0 {
			o.RetryWaitMax = c.RetryWaitMax
		}

		if c.RetryTimeout != 0 {
			o.RetryTimeout = c.RetryTimeout
		}

		if c.PooledClient != nil {
			o.PooledClient = *c.PooledClient
		}

		if len(c.Header) > 0 {
			o.Header = c.Header
		}

		if c.Proxy != "" {
			o.Proxy = c.Proxy
		}

		if c.HTTP2 != nil {
			o.HTTP2 = *c.HTTP2
		}

		if c.TLSConfig != nil {
			o.TLSConfig = c.TLSConfig
		}
	}
}

// New creates a new client with the configuration.
//   - Add pre defined options
func (c *Config) New(options ...OptionClientFn) (*Client, error) {
	return New(append(options, c.ToOption())...)
}
