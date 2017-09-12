package ultraclient

import (
	"math/rand"
	"net/url"
	"time"
)

// RandomStrategy implements Strategy for random endopoint selection
type RandomStrategy struct {
	endpoints []url.URL
	rand      *rand.Rand
}

// NextEndpoint returns an endpoint using a random strategy
func (r *RandomStrategy) NextEndpoint() url.URL {
	return r.endpoints[r.rand.Intn(len(r.endpoints))]
}

// SetEndpoints sets the available endpoints for use by the strategy
func (r *RandomStrategy) SetEndpoints(endpoints []url.URL) {
	s := rand.NewSource(time.Now().UnixNano())
	r.rand = rand.New(s)

	r.endpoints = endpoints
}

// GetEndpoints returns a random endpoint
func (r *RandomStrategy) GetEndpoints() []url.URL {
	return r.endpoints
}

// Length returns the number of endpoints
func (r *RandomStrategy) Length() int {
	return len(r.endpoints)
}

// Clone creates a clone of this strategy
func (r *RandomStrategy) Clone() LoadbalancingStrategy {
	rs := &RandomStrategy{}
	rs.SetEndpoints(r.endpoints)

	return rs
}
