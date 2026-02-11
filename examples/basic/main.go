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
The most basic example of a relay using rely.
*/

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	relay := rely.NewRelay()
	relay.On.Event = LogEvent
	relay.On.Req = LogReq

	if err := relay.StartAndServe(ctx, "localhost:3334"); err != nil {
		panic(err)
	}
}

func LogEvent(c rely.Client, e *nostr.Event) error {
	slog.Info("received event", "ID", e.ID, "IP", c.IP().Group())
	return nil
}

func LogReq(ctx context.Context, c rely.Client, id string, f nostr.Filters) ([]nostr.Event, error) {
	slog.Info("received req", "ID", id, "filters", len(f), "IP", c.IP().Group())
	return nil, nil
}
