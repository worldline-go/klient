package klient

import "io"

type optionDrain struct {
	Limit int64
}

func newOptionDrain(opts []OptionDrain) *optionDrain {
	o := new(optionDrain)
	for _, opt := range opts {
		opt(o)
	}

	if o.Limit == 0 {
		o.Limit = ResponseErrLimit
	}

	return o
}

type OptionDrain func(*optionDrain)

// WithDrainLimit sets the limit of the content to be read.
// If the limit is less than 0, it will read all the content.
func WithDrainLimit(limit int64) OptionDrain {
	return func(o *optionDrain) {
		o.Limit = limit
	}
}

// DrainBody reads the limited content of r and then closes the underlying io.ReadCloser.
func DrainBody(body io.ReadCloser, opts ...OptionDrain) {
	o := newOptionDrain(opts)

	defer body.Close()
	if o.Limit < 0 {
		_, _ = io.Copy(io.Discard, body)

		return
	}

	_, _ = io.Copy(io.Discard, io.LimitReader(body, o.Limit))
}
