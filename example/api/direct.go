package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/klient"
)

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
