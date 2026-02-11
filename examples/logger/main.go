package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pippellia-btc/rely"
)

/*
An example showcasing how to pass a custom logger for the relay to use.
*/

var logger *slog.Logger

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// creating a structured JSON logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	relay := rely.NewRelay(
		rely.WithLogger(logger),
	)

	relay.On.Event = LogEvent(logger)
	relay.On.Req = LogReq(logger)

	if err := relay.StartAndServe(ctx, "localhost:3334"); err != nil {
		panic(err)
	}
}

func LogEvent(logger *slog.Logger) func(c rely.Client, e *nostr.Event) error {
	return func(c rely.Client, e *nostr.Event) error {
		logger.Info("received event", "ID", e.ID, "IP", c.IP().Group())
		return nil
	}
}

func LogReq(logger *slog.Logger) func(ctx context.Context, c rely.Client, id string, f nostr.Filters) ([]nostr.Event, error) {
	return func(ctx context.Context, c rely.Client, id string, f nostr.Filters) ([]nostr.Event, error) {
		logger.Info("received req", "ID", id, "filters", len(f), "IP", c.IP().Group())
		return nil, nil
	}
}
