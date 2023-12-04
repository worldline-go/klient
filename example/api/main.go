package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/logz"
)

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

type GetBeerRequest struct {
	ID string
}

type GetBeerRespond struct {
	Name string `json:"name"`
}

// GetBeer directly initialize the request and response function.
func (c BeerAPI) GetBeer(ctx context.Context, request GetBeerRequest) (GetBeerRespond, error) {
	var v []GetBeerRespond

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("beers/%s", request.ID), nil)
	if err != nil {
		return GetBeerRespond{}, err
	}

	if err := c.client.Do(req, klient.ResponseFuncJSON(&v)); err != nil {
		return GetBeerRespond{}, err
	}

	return v[0], nil
}

func main() {
	logz.InitializeLog()

	client, err := klient.New(
		klient.WithBaseURL("https://api.punkapi.com/v2/"),
		// klient.WithBaseURL("https://expired.badssl.com/"),
		// klient.WithDisableEnvValues(true),
		// klient.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// 	Level: slog.LevelDebug,
		// }))),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}

	ctx := context.Background()

	beerAPI := BeerAPI{client}
	beer, err := beerAPI.GetRandomBeer(ctx)
	// beer, err := beerAPI.GetBeer(ctx, GetBeerRequest{ID: "1"})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get random beer")
	}

	log.Info().Interface("beer", beer).Msg("got beer")
}
