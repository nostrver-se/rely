package auth

import (
	"errors"
	"time"
)

type Config struct {
	// MaxPubkeys is the maximum number of public keys that can be authenticated at once.
	// Further attempts will be rejected with the error [ErrTooManyAuthed].
	// Default is 64.
	MaxPubkeys int

	// URL is the canonical relay URL (host + path) that must match the "relay" tag in AUTH events.
	// Examples: "relay.example.com", "example.com/relay"
	// Scheme and port are ignored as they are transport details.
	URL string

	// ChallengeBytes is the number of bytes in the challenge string.
	// Default is 16 bytes.
	ChallengeBytes uint8

	// TimeTolerance is the maximum allowed clock skew between the client and the server.
	// Default is 1 minute.
	TimeTolerance time.Duration
}

func NewConfig() Config {
	return Config{
		MaxPubkeys:     64,
		ChallengeBytes: 16,
		TimeTolerance:  time.Minute,
	}
}

func (c Config) Validate() error {
	if c.MaxPubkeys <= 0 {
		return errors.New("max pubkeys must be positive")
	}
	if c.ChallengeBytes <= 0 {
		return errors.New("challenge bytes must be positive. Suggested is 16 bytes")
	}
	if c.TimeTolerance <= 0 {
		return errors.New("time tolerance must be positive. Suggested is 1 minute")
	}
	return nil
}
