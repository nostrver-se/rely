package main

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pippellia-btc/rely"
)

/*
This programs shows how nostr events can be used as requests for short or long lived
asynchronous jobs like DVMs. The store is obviously a joke.
*/

var (
	store []*nostr.Event
	relay *rely.Relay
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	store = make([]*nostr.Event, 0, 1000)

	relay = rely.NewRelay(
		rely.WithQueueCapacity(10_000), // increase capacity to absorb traffic bursts (higher RAM)
		rely.WithMaxProcessors(10),     // increase concurrent processors for faster execution (higher CPU)
	)

	relay.Reject.Event.Append(
		KindNotIn([]int{5500}),
	)
	relay.On.Event = Process
	relay.On.Req = Query

	if err := relay.StartAndServe(ctx, "localhost:3334"); err != nil {
		panic(err)
	}
}

// KindNotIn returns a filter function that rejects events with a kind not in the given list.
func KindNotIn(kinds []int) func(rely.Client, *nostr.Event) error {
	return func(_ rely.Client, e *nostr.Event) error {
		if slices.Contains(kinds, e.Kind) {
			return nil
		}
		return errors.New("unsupported kind")
	}
}

func Process(_ rely.Client, request *nostr.Event) error {
	// malware scanning DVM
	response := MalwareScan(request)

	// store so it can later be retrieved
	store = append(store, response)

	// broadcast the response to all clients right away
	relay.Broadcast(response)

	// we don't need to save the request
	return nil
}

func Query(ctx context.Context, _ rely.Client, _ string, filters nostr.Filters) ([]nostr.Event, error) {
	events := make([]nostr.Event, 0, 100) // pre-allocating
	for _, event := range store {
		if filters.Match(event) {
			events = append(events, *event)
		}
	}
	return events, nil
}

func MalwareScan(request *nostr.Event) *nostr.Event {
	switch rand.IntN(3) {
	case 0:
		// no malware detected
		return responseEvent(request, true)

	case 1:
		// malware detected
		return responseEvent(request, false)

	default:
		// an error occurred
		return errorEvent(request)
	}
}

func responseEvent(request *nostr.Event, isMalware bool) *nostr.Event {
	content := "all good"
	if isMalware {
		content = "found virus"
	}

	return &nostr.Event{
		Content: content,
		Kind:    request.Kind + 1000,
		Tags:    nostr.Tags{{"e", request.ID}, {"p", request.PubKey}},
	}
}

func errorEvent(request *nostr.Event) *nostr.Event {
	return &nostr.Event{
		Kind: 7000,
		Tags: nostr.Tags{{"e", request.ID}, {"p", request.PubKey}, {"status", "whatever"}},
	}
}
