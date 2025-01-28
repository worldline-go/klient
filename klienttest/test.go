package klienttest

import (
	"net/http"

	"github.com/worldline-go/klient"
)

type KlientTest struct {
	client    *klient.Client
	transport *TransportHandler
}

func New(opts ...klient.OptionClientFn) (*KlientTest, error) {
	k := &KlientTest{
		transport: &TransportHandler{
			Handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"message": "using default klienttest handler"}`))
			},
		},
	}

	client, err := klient.New(append(opts, klient.WithBaseTransport(k.transport))...)
	if err != nil {
		return nil, err
	}

	k.client = client

	return k, nil
}

func (k *KlientTest) SetHandler(handler http.HandlerFunc) {
	k.transport.SetHandler(handler)
}

func (k *KlientTest) Client() *klient.Client {
	return k.client
}
