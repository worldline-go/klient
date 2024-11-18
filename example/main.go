package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/klient/example/beer"
	"github.com/worldline-go/logz"
)

func BeerExample(ctx context.Context) error {
	api, err := beer.New()
	if err != nil {
		return err
	}

	beerResult, err := api.GetRandomBeer(ctx)
	// beer, err := beerAPI.GetBeer(ctx, GetBeerRequest{ID: "5128df48-79fc-4f0f-8b52-d06be54d0cec"})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get random beer")
	}

	log.Info().Interface("beer", beerResult).Msg("got beer")

	return nil
}

func HTTP2Example(ctx context.Context) error {
	client, err := klient.New(
		klient.WithHTTP2(true),
		klient.WithHeaderSet(http.Header{
			"Content-Type": []string{"application/json"},
			"X-User":       []string{"test-user"},
		}),
		klient.WithBaseURL("http://localhost:8080"),
	)
	if err != nil {
		return err
	}

	// https://http2.golang.org/
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)
	if err != nil {
		return err
	}

	if err := client.Do(req, func(r *http.Response) error {
		log.Info().Str("proto", r.Proto).Msg("got response")

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func main() {
	logz.InitializeLog()

	if err := run(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to run")
	}
}

func run(ctx context.Context) error {
	switch v := os.Getenv("EXAMPLE"); strings.ToUpper(v) {
	case "BEER":
		return BeerExample(ctx)
	case "HTTP2":
		return HTTP2Example(ctx)
	default:
		return fmt.Errorf("unknown 'EXAMPLE' env: %s", v)
	}
}
