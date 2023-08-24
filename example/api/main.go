package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/logz"
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

func (r RandomRequest) Request(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, r.Method(), r.Path(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

type RandomRequestResponse struct {
	Name string `json:"name"`
}

func (c BeerAPI) GetRandomBeer(ctx context.Context) ([]RandomRequestResponse, error) {
	var v []RandomRequestResponse

	if err := c.klient.DoWithInf(ctx, RandomRequest{}, func(r *http.Response) error {
		if r.StatusCode != http.StatusOK {
			return klient.ResponseError(r)
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

func (r GetBeerRequest) Request(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, r.Method(), r.Path(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

type GetBeerRespond struct {
	Name string `json:"name"`
}

func (c BeerAPI) GetBeer(ctx context.Context, req GetBeerRequest) (GetBeerRespond, error) {
	var v []GetBeerRespond

	if err := c.klient.DoWithInf(ctx, req, klient.ResponseFuncJSON(&v)); err != nil {
		return GetBeerRespond{}, err
	}

	return v[0], nil
}

func DirectCall(ctx context.Context, client *klient.Client) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "beers/random", nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create request")
	}

	var response interface{}

	if err := client.Do(req, func(r *http.Response) error {
		if r.StatusCode != http.StatusOK {
			return klient.ResponseError(r)
		}

		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w, body: %v", err, klient.LimitedResponse(r))
		}

		return nil
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to request")
	}

	log.Info().Interface("beers", response).Msg("got beers")
}

func main() {
	logz.InitializeLog()

	client, err := klient.New(
		klient.OptionClient.WithBaseURL("https://api.punkapi.com/v2/"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}

	ctx := context.Background()

	// DirectCall(ctx, client)
	// or
	beerAPI := BeerAPI{client}
	beer, err := beerAPI.GetRandomBeer(ctx)
	// beer, err := beerAPI.GetBeer(ctx, GetBeerRequest{ID: "1"})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get random beer")
	}

	log.Info().Interface("beer", beer).Msg("got beer")
}
