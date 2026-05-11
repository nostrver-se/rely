package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pippellia-btc/rely"
)

/*
This example presents a reputation-based rate limiter using a two-tier token bucket system:
- client must authenticate with exactly one pubkey (via NIP-42)
- each pubkey has an associate bucket holding tokens
- writing an event comsumes a token
- tokens are refilled periodically based on the pubkey's rank (global pagerank)
- ranks are fetched in batches from a service provider (Vertex)
- to prevent abuse (e.g. attackers forcing mass ranking of pubkeys), IP-based rate
limiting is applied to the ranking process

Note: Unknown pubkeys are treated as having zero reputation,
while unknown IPs are initially given a positive reputation.

Source: https://vertexlab.io/blog/reputation_rate_limit
*/

const relayBudget = 10_000_000 // maximum events per day
const ipTokens = 100

var cache *RankCache
var limiter *Limiter

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cache = NewRankCache(ctx)
	limiter = NewLimiter(ctx)

	relay := rely.NewRelay(
		rely.WithAuthURL("relay.example.com"), // required for validating NIP-42 auth
		rely.WithoutMultiAuth(),               // enforcing max one pubkey per client
	)

	// apply connection-level ip rate limiting
	relay.Reject.Connection.Prepend(RateLimitConn)

	// send an AUTH challenge as soon as the client connects
	relay.On.Connect = func(c rely.Client) { c.SendAuth() }

	// reject events from clients that are not authenticated, or have exhausted their pubkey rate limit
	relay.Reject.Event.Append(
		NotAuthed,
		RateLimitPubkey,
	)

	relay.On.Event = func(c rely.Client, e *nostr.Event) error {
		if err := Save(e); err != nil {
			return err
		}
		return nil
	}

	if err := relay.StartAndServe(ctx, "localhost:3334"); err != nil {
		panic(err)
	}
}

// RateLimitConn is a connection-level rate limiter.
func RateLimitConn(_ rely.Stats, r *http.Request) error {
	// rate limiting IPs
	ip := rely.GetIP(r).Group()
	if limiter.Allow(ip, ipRefill) {
		return nil
	}
	return errors.New("rate-limited: please try again in a few hours")
}

func NotAuthed(c rely.Client, _ *nostr.Event) error {
	if c.IsAuthed() {
		return nil
	}
	return errors.New("auth-required: you must be authenticated with exactly one pubkey to write here")
}

func RateLimitPubkey(c rely.Client, _ *nostr.Event) error {
	pubkey := c.Pubkeys()[0]
	rank, found := cache.Rank(pubkey)
	if !found {
		// If the client IP has enough tokens, the pubkey is queued for ranking by Vertex;
		// otherwise we disconnect the client as this is probably an attacker trying to waste our backend budget.
		if !limiter.Allow(c.IP().Group(), ipRefill) {
			c.Disconnect()
			return errors.New("rate-limited: please try again in a few hours")
		}

		cache.refresh <- pubkey
	}

	if !limiter.Allow(pubkey, pkRefill(rank)) {
		return errors.New("rate-limited: please try again in a few hours")
	}
	return nil
}

func Save(e *nostr.Event) error {
	log.Println(e)
	return nil
}
