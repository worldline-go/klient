package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/klient"
)

func main() {
	client, err := klient.NewClient(klient.OptionClient.WithBaseURL("https://api.punkapi.com/v2/"))
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
