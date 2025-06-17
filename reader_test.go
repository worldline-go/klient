package klient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

func TestReader(t *testing.T) {
	t.Run("concat reader", func(t *testing.T) {
		data := []byte(`
	Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
	Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
	Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
	Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

	Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
	Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
	Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
	Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
	`)
		readerData := bytes.NewReader(data)

		// read part of the data
		partData, err := io.ReadAll(io.LimitReader(readerData, 5))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// merge 2 readers together
		r := NewMultiReader(io.NopCloser(bytes.NewReader(partData)), io.NopCloser(readerData))

		// read the rest of the data
		allData, err := io.ReadAll(r)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if string(allData) != string(data) {
			t.Errorf("expected %s, got %s", string(data), string(allData))
		}

		if err := r.Close(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("context cancel", func(t *testing.T) {
		data := []byte("Hello, World!")
		readerData := bytes.NewReader(data)

		// read part of the data
		partData, _ := io.ReadAll(io.LimitReader(readerData, 5))
		// merge 2 readers together
		r := NewMultiReader(io.NopCloser(bytes.NewReader(partData)), io.NopCloser(readerData))

		ctx, cancel := context.WithCancel(t.Context())
		r.SetContext(ctx)
		cancel()

		// read the rest of the data
		_, err := io.ReadAll(r)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("small parts", func(t *testing.T) {
		data1 := []byte("Hello")
		data2 := []byte(", World!")

		r := NewMultiReader(io.NopCloser(bytes.NewReader(data1)), io.NopCloser(bytes.NewReader(data2)))

		p := make([]byte, 0, 50)
		n, err := r.Read(p[len(p):cap(p)])
		if !errors.Is(err, io.EOF) {
			t.Errorf("unexpected error: %v", err)
		}
		p = p[:n]

		if lenDatas := (len(data1) + len(data2)); n != lenDatas {
			t.Errorf("expected %d, got %d", lenDatas, n)
		}

		if string(p) != "Hello, World!" {
			t.Errorf("expected Hello, got %s", string(p))
		}

		if len(p) != 13 {
			t.Errorf("expected 13, got %d", len(p))
		}
	})
}
