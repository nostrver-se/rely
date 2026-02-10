package auth

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/nbd-wtf/go-nostr"
	"github.com/pippellia-btc/smallset"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name     string
		event    *nostr.Event
		expected error
	}{
		{
			name:     "invalid kind",
			event:    Signed(nostr.Event{Kind: 69, ID: "abc", CreatedAt: nostr.Now()}),
			expected: ErrInvalidKind,
		},
		{
			name:     "no relay tag",
			event:    Signed(nostr.Event{Kind: 22242, CreatedAt: nostr.Now(), Tags: nostr.Tags{{"challenge", "challenge"}}}),
			expected: ErrInvalidRelay,
		},
		{
			name:     "no challenge tag",
			event:    Signed(nostr.Event{Kind: 22242, CreatedAt: nostr.Now(), Tags: nostr.Tags{{"relay", "example.com"}}}),
			expected: ErrInvalidChallenge,
		},
		{
			name:     "invalid ID",
			event:    &nostr.Event{Kind: 22242, CreatedAt: nostr.Now(), Tags: nostr.Tags{{"challenge", "challenge"}, {"relay", "example.com"}}},
			expected: ErrInvalidEventID,
		},
		{
			name:     "invalid signature",
			event:    WithBadSignature(nostr.Event{Kind: 22242, CreatedAt: nostr.Now(), Tags: nostr.Tags{{"relay", "example.com"}, {"challenge", "challenge"}}}),
			expected: ErrInvalidEventSignature,
		},
		{
			name:  "valid",
			event: Signed(nostr.Event{Kind: 22242, CreatedAt: nostr.Now(), Tags: nostr.Tags{{"relay", "example.com"}, {"challenge", "challenge"}}}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(test.event); err != nil {
				t.Fatalf("failed to encode event: %v", err)
			}

			_, err := Parse(json.NewDecoder(&buf))
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected error %v, got %v", test.expected, err)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	canonical, _ := CanonicalURL("example.com")
	state := State{
		pubkeys:   smallset.New[string](10),
		challenge: "challenge",
		config: Config{
			URL:           canonical,
			TimeTolerance: time.Minute,
		},
	}

	tests := []struct {
		name     string
		request  Request
		expected error
	}{
		{
			name:     "too much into the past",
			request:  Request{ID: "id-1", Pubkey: "pubkey-1", CreatedAt: time.Now().Add(-2 * time.Minute), Challenge: "challenge", Relay: "example.com"},
			expected: ErrInvalidTimestamp,
		},
		{
			name:     "relay is different",
			request:  Request{ID: "id-2", Pubkey: "pubkey-2", CreatedAt: time.Now(), Challenge: "challenge", Relay: "example.com.evil.website"},
			expected: ErrInvalidRelay,
		},
		{
			name:     "challenge is different",
			request:  Request{ID: "id-3", Pubkey: "pubkey-3", CreatedAt: time.Now(), Challenge: "different", Relay: "example.com"},
			expected: ErrInvalidChallenge,
		},
		{
			name:     "valid",
			request:  Request{ID: "id-4", Pubkey: "pubkey-4", CreatedAt: time.Now(), Challenge: "challenge", Relay: "example.com"},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := state.Validate(test.request)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected error %v, got %v", test.expected, err)
			}
		})
	}
}

func Signed(e nostr.Event) *nostr.Event {
	sk := nostr.GeneratePrivateKey()
	e.Sign(sk)
	return &e
}

func WithBadSignature(e nostr.Event) *nostr.Event {
	ev := Signed(e)
	ev.Sig = "bad"
	return ev
}

func TestCanonicalURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Basic cases
		{
			name:     "simple domain",
			input:    "example.com",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "domain with wss scheme",
			input:    "wss://example.com",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "domain with ws scheme",
			input:    "ws://example.com",
			expected: "example.com",
			wantErr:  false,
		},
		// Path cases
		{
			name:     "domain with path",
			input:    "example.com/relay",
			expected: "example.com/relay",
			wantErr:  false,
		},
		{
			name:     "domain with path and scheme",
			input:    "wss://example.com/relay",
			expected: "example.com/relay",
			wantErr:  false,
		},
		{
			name:     "domain with path and trailing slash",
			input:    "wss://example.com/relay/",
			expected: "example.com/relay",
			wantErr:  false,
		},
		{
			name:     "domain with nested path",
			input:    "wss://example.com/nostr/relay",
			expected: "example.com/nostr/relay",
			wantErr:  false,
		},
		// Port cases (should be ignored)
		{
			name:     "domain with standard wss port",
			input:    "wss://example.com:443",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "domain with custom port",
			input:    "wss://example.com:8080",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "domain with port and path",
			input:    "wss://example.com:8080/relay",
			expected: "example.com/relay",
			wantErr:  false,
		},
		{
			name:     "ws with port 80",
			input:    "ws://example.com:80/relay",
			expected: "example.com/relay",
			wantErr:  false,
		},
		// Case normalization
		{
			name:     "uppercase domain",
			input:    "wss://Example.COM",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "mixed case domain with path",
			input:    "wss://Example.Com/Relay",
			expected: "example.com/Relay",
			wantErr:  false,
		},
		// IPv6 cases
		{
			name:     "ipv6 localhost",
			input:    "wss://[::1]",
			expected: "::1",
			wantErr:  false,
		},
		{
			name:     "ipv6 with port",
			input:    "wss://[::1]:7777",
			expected: "::1",
			wantErr:  false,
		},
		{
			name:     "ipv6 with path",
			input:    "wss://[2001:db8::1]/relay",
			expected: "2001:db8::1/relay",
			wantErr:  false,
		},
		// Whitespace handling
		{
			name:     "leading and trailing whitespace",
			input:    "  wss://example.com/relay  ",
			expected: "example.com/relay",
			wantErr:  false,
		},
		// Root path handling
		{
			name:     "root path with trailing slash",
			input:    "wss://example.com/",
			expected: "example.com",
			wantErr:  false,
		},
		// Error cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid URL",
			input:    "://invalid",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "userinfo in URL (security)",
			input:    "wss://user:pass@example.com",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "userinfo without password",
			input:    "wss://user@example.com",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "userinfo with path",
			input:    "wss://user@example.com/relay",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "scheme only",
			input:    "wss://",
			expected: "",
			wantErr:  true,
		},
		// Real-world examples
		{
			name:     "nostr.band main relay",
			input:    "wss://relay.nostr.band",
			expected: "relay.nostr.band",
			wantErr:  false,
		},
		{
			name:     "nostr.band sub-relay with path",
			input:    "wss://relay.nostr.band/all",
			expected: "relay.nostr.band/all",
			wantErr:  false,
		},
		{
			name:     "damus relay",
			input:    "wss://relay.damus.io",
			expected: "relay.damus.io",
			wantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := CanonicalURL(test.input)

			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	state := State{
		pubkeys:   smallset.New[string](10),
		challenge: "test-challenge",
		config: Config{
			URL:           "relay.example.com/nostr",
			TimeTolerance: time.Minute,
		},
	}

	tests := []struct {
		name      string
		relayURL  string
		shouldErr bool
	}{
		{
			name:      "exact match",
			relayURL:  "wss://relay.example.com/nostr",
			shouldErr: false,
		},
		{
			name:      "different scheme (ws)",
			relayURL:  "ws://relay.example.com/nostr",
			shouldErr: false,
		},
		{
			name:      "different port",
			relayURL:  "wss://relay.example.com:8080/nostr",
			shouldErr: false,
		},
		{
			name:      "different case",
			relayURL:  "wss://Relay.Example.COM/nostr",
			shouldErr: false,
		},
		{
			name:      "with trailing slash",
			relayURL:  "wss://relay.example.com/nostr/",
			shouldErr: false,
		},
		{
			name:      "different path",
			relayURL:  "wss://relay.example.com/other",
			shouldErr: true,
		},
		{
			name:      "different host",
			relayURL:  "wss://other.example.com/nostr",
			shouldErr: true,
		},
		{
			name:      "with userinfo (should fail)",
			relayURL:  "wss://user@relay.example.com/nostr",
			shouldErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := Request{
				ID:        "test-id",
				Pubkey:    "test-pubkey",
				CreatedAt: time.Now(),
				Challenge: "test-challenge",
				Relay:     test.relayURL,
			}

			err := state.Validate(req)
			hasErr := (err != nil)

			if hasErr != test.shouldErr {
				t.Fatalf("expected error=%v, got error=%v", test.shouldErr, err)
			}
		})
	}
}
