package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/pippellia-btc/rely"
)

/*
The most basic example of a relay using rely.
*/

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	/* Add NIP-11 */
	relayInfo := nip11.RelayInformationDocument{
    Name:           "rely.nyves.nl",
    Description:    "",
    PubKey:         "9919981db5d9f59713bd717069212c32a19b55f2280a51e70c49c5b73da306b9",
    SupportedNIPs:  []any{1, 11, 42, 45},
    RelayCountries: []string{"Netherlands"},
  }

	/* @todo Add NIP-50 */

	relay := rely.NewRelay(
	  rely.WithAuthURL("rely.nyves.nl"),
	  rely.WithInfo(relayInfo),
	)
  relay.On.Count = Count
	relay.On.Event = LogEvent
	relay.On.Req = LogReq

	if err := relay.StartAndServe(ctx, "localhost:2013"); err != nil {
		panic(err)
	}
}

/* NIP-45 Count */
func Count(c rely.Client, id string, f nostr.Filters) (count int64, approx bool, err error) {
	slog.Info("received count", "id", id, "filters", f)
	count = rand.Int64N(10000)
	return count, (count % 2) == 1, nil
}

func LogEvent(c rely.Client, e *nostr.Event) error {
	slog.Info("received event", "id", e.ID, "ip", c.IP().Group())
	return nil
}

func LogReq(ctx context.Context, c rely.Client, id string, f nostr.Filters) ([]nostr.Event, error) {
	slog.Info("received req", "id", id, "filters", len(f), "ip", c.IP().Group())
	return nil, nil
}
