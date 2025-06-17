package beer

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/worldline-go/klient"
	"github.com/worldline-go/klient/klienttest"
)

func TestBeerAPI_GetBeer(t *testing.T) {
	type args struct {
		request GetBeerRequest
	}
	tests := []struct {
		name    string
		handler http.HandlerFunc
		args    args
		want    GetBeerRespond
		wantErr bool
	}{
		{
			name: "call",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Scheme != "https" ||
					r.URL.Host != "api.openbrewerydb.org" ||
					r.URL.Path != "/v1/breweries/1" ||
					r.Header.Get("Content-Type") != "application/json" ||
					r.Method != http.MethodGet {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error": "not found"}`))

					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[{"name":"test"}]`))
			},
			args: args{
				request: GetBeerRequest{
					ID: "1",
				},
			},
			want: GetBeerRespond{
				Name: "test",
			},
			wantErr: false,
		},
	}

	transport := new(klienttest.TransportHandler)
	c, err := New(BeerAPIConfig.ToOption(), klient.WithBaseTransport(transport))
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport.SetHandler(tt.handler)

			got, err := c.GetBeer(t.Context(), tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeerAPI.GetBeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeerAPI.GetBeer() = %v, want %v", got, tt.want)
			}
		})
	}
}
