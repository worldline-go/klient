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

func DirectCall(ctx context.Context, client *klient.Client) {
	if err := client.Request(ctx, http.MethodGet, "beers/random", nil, nil, func(r *http.Response) error {
		if r.StatusCode != http.StatusOK {
			return klient.UnexpectedResponseError(r)
		}

		var response interface{}

		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w, body: %v", err, klient.LimitedResponse(r))
		}

		log.Info().Interface("beers", response).Msg("got beers")

		return nil
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to request")
	}
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
	// beer, err := beerAPI.GetRandomBeer(ctx)
	beer, err := beerAPI.GetBeer(ctx, GetBeerRequest{ID: "1"})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get random beer")
	}

	log.Info().Interface("beer", beer).Msg("got beer")
}
