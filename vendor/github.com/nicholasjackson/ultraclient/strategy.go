package ultraclient

import (
	"net/url"
	"time"
)

// LoadbalancingStrategy is an interface to be implemented by loadbalancing
// strategies like round robin or random.
type LoadbalancingStrategy interface {
	// NextEndpoint returns the next endpoint in the strategy
	NextEndpoint() url.URL

	// SetEndpoints sets or updates the endpoints for the strategy
	SetEndpoints([]url.URL)

	// GetEndpoints returns the endpoints for the strategy
	GetEndpoints() []url.URL

	// Lenght returns the number of endpoints
	Length() int

	// Clone creates a deep clone of this object, rather than passing the main
	// instance to a client clone should be called which ensures that the
	// client obtains the correct behaviour for the loadbalancer.
	Clone() LoadbalancingStrategy
}

// BackoffStrategy implements a strategy for retry backoffs
type BackoffStrategy interface {
	Create(retries int, delay time.Duration) []time.Duration
}
