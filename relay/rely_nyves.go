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

	/* @todo Add NIP-11 */

	/* @todo Add NIP-45 */

	/* @todo Add NIP-50 */

	relay := rely.NewRelay()
	relay.On.Event = LogEvent
	relay.On.Req = LogReq

	if err := relay.StartAndServe(ctx, "localhost:2013"); err != nil {
		panic(err)
	}
}

func LogEvent(c rely.Client, e *nostr.Event) error {
	slog.Info("received event", "id", e.ID, "ip", c.IP().Group())
	return nil
}

func LogReq(ctx context.Context, c rely.Client, id string, f nostr.Filters) ([]nostr.Event, error) {
	slog.Info("received req", "id", id, "filters", len(f), "ip", c.IP().Group())
	return nil, nil
}
