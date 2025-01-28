package klienttest

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

// TransportHandler is base of http.RoundTripper.
type TransportHandler struct {
	Handler http.HandlerFunc

	m sync.RWMutex
}

var _ http.RoundTripper = (*TransportHandler)(nil)

func (t *TransportHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	t.m.RLock()
	defer t.m.RUnlock()

	recorder := httptest.NewRecorder()

	if t.Handler != nil {
		t.Handler(recorder, req)
	}

	return recorder.Result(), nil
}

func (t *TransportHandler) SetHandler(handler http.HandlerFunc) {
	t.m.Lock()
	defer t.m.Unlock()

	t.Handler = handler
}
