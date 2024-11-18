package beer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/worldline-go/klient"
)

type BeerAPI struct {
	client *klient.Client
}

func New(opts ...klient.OptionClientFn) (*BeerAPI, error) {
	client, err := klient.Config{
		BaseURL: "https://api.openbrewerydb.org/v1/",
	}.New(opts...)
	if err != nil {
		return nil, err
	}

	return &BeerAPI{client}, nil
}

type RandomGet struct{}

func (r RandomGet) Request(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "breweries/random", nil)
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

type GetBeerRequest struct {
	ID string
}

type GetBeerRespond struct {
	Name string `json:"name"`
}

// GetBeer directly initialize the request and response function.
func (c BeerAPI) GetBeer(ctx context.Context, request GetBeerRequest) (GetBeerRespond, error) {
	var v []GetBeerRespond

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("breweries/%s", request.ID), nil)
	if err != nil {
		return GetBeerRespond{}, err
	}

	if err := c.client.Do(req, klient.ResponseFuncJSON(&v)); err != nil {
		return GetBeerRespond{}, err
	}

	return v[0], nil
}
