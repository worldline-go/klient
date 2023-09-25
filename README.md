# klient

[![License](https://img.shields.io/github/license/worldline-go/klient?color=red&style=flat-square)](https://raw.githubusercontent.com/worldline-go/klient/main/LICENSE)
[![Coverage](https://img.shields.io/sonar/coverage/worldline-go_klient?logo=sonarcloud&server=https%3A%2F%2Fsonarcloud.io&style=flat-square)](https://sonarcloud.io/summary/overall?id=worldline-go_klient)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/worldline-go/klient/test.yml?branch=main&logo=github&style=flat-square&label=ci)](https://github.com/worldline-go/klient/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/worldline-go/klient?style=flat-square)](https://goreportcard.com/report/github.com/worldline-go/klient)
[![Go PKG](https://raw.githubusercontent.com/worldline-go/guide/main/badge/custom/reference.svg)](https://pkg.go.dev/github.com/worldline-go/klient)


Retryable http client with some helper functions.

```sh
go get github.com/worldline-go/klient
```

## Usage

Create a new client with a base url. Base url is mandatory in default also it can set with `API_GATEWAY_ADDRESS` environment variable.

```go
client, err := klient.New(klient.OptionClient.WithBaseURL("https://api.punkapi.com/v2/"))
if err != nil {
    // handle error
}
```

Client has `Do` and `DoWithInf` methods to send request.  
Methods will automatically drain and close the response body and it resolves reference of URL.

### Request with http.Request

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, "beers/random", nil)
if err != nil {
	// handle error
}

var response interface{}

if err := client.Do(req, func(r *http.Response) error {
	if r.StatusCode != http.StatusOK {
		return klient.ErrResponse(r)
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w, body: %v", err, klient.LimitedResponse(r))
	}

	return nil
}); err != nil {
	// handle error
}

log.Info().Interface("beers", response).Msg("got beers")
```

### Request with interface

Our interface just one function to create a request.

```go
type Requester interface {
	Request(context.Context) (*http.Request, error)
}
```

Set an API's struct with has client.

```go
type BeerAPI struct {
	client *klient.Client
}

type RandomGet struct{}

func (r RandomGet) Request(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "beers/random", nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (r RandomGet) Response(resp *http.Response) ([]RandomRequestResponse, error) {
	var v []RandomRequestResponse
	if err := klient.ResponseFuncJSON(&v)(resp); err != nil {
		return nil, err
	}

	return v, nil
}


type RandomRequestResponse struct {
	Name string `json:"name"`
}

func (c BeerAPI) GetRandomBeer(ctx context.Context) ([]RandomRequestResponse, error) {
	v, err := klient.DoWithInf(ctx, c.client.HTTP, RandomGet{})
	if err != nil {
		return nil, err
	}

	return v, nil
}
```

Now you need to create a new instance of your API and use it.

```go
api := BeerAPI{
    klient: client,
}

respond, err := api.GetRandomBeer(ctx)
if err != nil {
    // handle error
}
```

## Env values

| Name                          | Description                                                           |
| ----------------------------- | --------------------------------------------------------------------- |
| `API_GATEWAY_ADDRESS`         | Base url of client if not set with `WithBaseURL`.                     |
| `KLIENT_BASE_URL`             | Base url of client same with `API_GATEWAY_ADDRESS` but more priority. |
| `KLIENT_INSECURE_SKIP_VERIFY` | Skip tls verify. Ex `KLIENT_INSECURE_SKIP_VERIFY=true`                |
| `KLIENT_TIMEOUT`              | Timeout for http client. Ex: `KLIENT_TIMEOUT=30s`                     |
