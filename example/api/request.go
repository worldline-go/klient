package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/worldline-go/klient"
)

type BeerAPI struct {
	klient *klient.Client
}

type RandomRequest struct{}

func (RandomRequest) Method() string {
	return http.MethodGet
}

func (RandomRequest) Path() string {
	return "beers/random"
}

type RandomRequestResponse struct {
	Name string `json:"name"`
}

func (c BeerAPI) GetRandomBeer(ctx context.Context) ([]RandomRequestResponse, error) {
	var v []RandomRequestResponse

	if err := c.klient.DoWithFunc(ctx, RandomRequest{}, func(r *http.Response) error {
		if r.StatusCode != http.StatusOK {
			return klient.UnexpectedResponseError(r)
		}

		return json.NewDecoder(r.Body).Decode(&v)
	}); err != nil {
		return nil, err
	}

	return v, nil
}

type GetBeerRequest struct {
	ID string
}

func (GetBeerRequest) Method() string {
	return http.MethodGet
}

func (r GetBeerRequest) Path() string {
	return fmt.Sprintf("beers/%s", r.ID)
}

type GetBeerRespond struct {
	Name string `json:"name"`
}

func (c BeerAPI) GetBeer(ctx context.Context, req GetBeerRequest) (GetBeerRespond, error) {
	var v []GetBeerRespond

	if err := c.klient.Do(ctx, req, &v); err != nil {
		return GetBeerRespond{}, err
	}

	return v[0], nil
}
