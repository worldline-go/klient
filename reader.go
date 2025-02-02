package klient

import (
	"context"
	"errors"
	"io"
)

type MultiReader struct {
	ctx context.Context
	rs  []io.ReadCloser
}

var _ io.ReadCloser = (*MultiReader)(nil)

// NewMultiReader returns a new read closer that reads from all the readers.
//   - This helps read small amount of body and concat read data with remains io.ReadCloser.
func NewMultiReader(rs ...io.ReadCloser) *MultiReader {
	return &MultiReader{rs: rs}
}

func (r *MultiReader) SetContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *MultiReader) Read(p []byte) (int, error) {
	nTotal, pTotal := 0, len(p)

	index := 0
	for {
		if r.ctx != nil && r.ctx.Err() != nil {
			return nTotal, r.ctx.Err()
		}

		if index >= len(r.rs) {
			return nTotal, io.EOF
		}

		rr := r.rs[index]

		n, err := rr.Read(p[nTotal:])
		nTotal += n
		pTotal -= n
		if pTotal == 0 {
			return nTotal, err
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nTotal, err
			}

			index++
		}
	}
}

func (r *MultiReader) Close() error {
	var err error
	for _, rr := range r.rs {
		if e := rr.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}
