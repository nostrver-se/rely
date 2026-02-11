package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/pippellia-btc/rely"
	. "github.com/pippellia-btc/rely"
)

/*
This example shows how to configure NIP-11 relay information document.
*/

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	info := nip11.RelayInformationDocument{
		Name:           "Rely",
		Description:    "this is just an example",
		PubKey:         "f683e87035f7ad4f44e0b98cfbd9537e16455a92cd38cefc4cb31db7557f5ef2",
		SupportedNIPs:  []any{1, 11, 42},
		RelayCountries: []string{"Italy"},
	}

	relay := NewRelay(
		WithInfo(info),
	)

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
