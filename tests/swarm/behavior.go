package swarm

import (
	"errors"
	"math"
	"math/rand/v2"

	"github.com/goccy/go-json"
)

// Behavior defines the behavior of a client of the swarm.
type Behavior interface {
	// NextRequest returns the bytes of the next request to send to the relay.
	NextRequest() []byte

	// ValidateResponse validates the response from the relay.
	ValidateResponse(*json.Decoder) error
}

// BehaviorDistribution holds a probability distribution of behaviors for swarm clients.
type BehaviorDistribution []struct {
	P        float64
	Behavior Behavior
}

func (b BehaviorDistribution) Validate() error {
	if len(b) == 0 {
		return errors.New("behavior distribution must not be empty")
	}

	total := 0.0
	for _, wb := range b {
		if wb.P < 0 || wb.P > 1 {
			return errors.New("probabilities must be between 0 and 1")
		}
		total += wb.P
	}
	if math.Abs(total-1) > 1e-5 {
		return errors.New("probabilities must sum to 1")
	}
	return nil
}

func (b BehaviorDistribution) Sample() Behavior {
	r := rand.Float64()
	for _, wb := range b {
		r -= wb.P
		if r < 0 {
			return wb.Behavior
		}
	}
	return b[len(b)-1].Behavior // safe fallback
}
