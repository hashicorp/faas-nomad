package ultraclient

import (
	"math/rand"
	"net/url"
	"time"
)

// RoundRobinStrategy is a load balancing strategy that implements a roundrobin
// style, it starts from a random location and the runs through each item in
// the endpoints collection sequentially
type RoundRobinStrategy struct {
	endpoints    []url.URL
	currentIndex int
}

// NextEndpoint returns an endpoint using a random strategy
func (r *RoundRobinStrategy) NextEndpoint() url.URL {
	r.currentIndex++
	if r.currentIndex >= len(r.endpoints) {
		r.currentIndex = 0
	}

	return r.endpoints[r.currentIndex]
}

// SetEndpoints sets the available endpoints for use by the strategy
func (r *RoundRobinStrategy) SetEndpoints(endpoints []url.URL) {
	s := rand.NewSource(time.Now().UnixNano())
	ra := rand.New(s)
	r.currentIndex = ra.Intn(len(endpoints))

	r.endpoints = endpoints
}

// GetEndpoints returns the next endpoint in the list
func (r *RoundRobinStrategy) GetEndpoints() []url.URL {
	return r.endpoints
}

// Length returns the nuimber of endpoints
func (r *RoundRobinStrategy) Length() int {
	return len(r.endpoints)
}

// Clone creates a clone of this strategy this should be called when creating
// a new client
func (r *RoundRobinStrategy) Clone() LoadbalancingStrategy {
	rs := &RoundRobinStrategy{}
	rs.SetEndpoints(r.endpoints)

	return rs
}
