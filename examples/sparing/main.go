package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pippellia-btc/rely"
)

/*
A "sparing" relay that avoids responding to REQs when the client's
response buffer is nearly full. This helps prevent overwhelming
slow clients and demonstrates how to use Client.RemainingCapacity
to apply simple backpressure.
*/

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	relay := rely.NewRelay(
		rely.WithClientResponseLimit(100), // decreased from default 1000
	)

	relay.Reject.Req.Prepend(TooGreedy)
	relay.On.Event = LogEvent
	relay.On.Req = LogReq

	if err := relay.StartAndServe(ctx, "localhost:3334"); err != nil {
		panic(err)
	}
}

func TooGreedy(client rely.Client, id string, filters nostr.Filters) error {
	if client.RemainingCapacity() < 10 {
		return errors.New("slow down there chief")
	}
	return nil
}

func LogEvent(c rely.Client, e *nostr.Event) error {
	slog.Info("received event", "id", e.ID, "ip", c.IP().Group())
	return nil
}

func LogReq(ctx context.Context, c rely.Client, id string, f nostr.Filters) ([]nostr.Event, error) {
	slog.Info("received req", "id", id, "filters", len(f), "ip", c.IP().Group())
	return nil, nil
}
