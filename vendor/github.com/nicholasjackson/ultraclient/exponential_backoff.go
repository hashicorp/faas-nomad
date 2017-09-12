package ultraclient

import (
	"time"

	"github.com/eapache/go-resiliency/retrier"
)

// ExponentialBackoff is a backoffStrategy which implements an exponential
// retry policy
type ExponentialBackoff struct {
	cache []time.Duration
}

// Create creates a new ExponentialBackoff timings with the given retries and
// iniital delay
func (e *ExponentialBackoff) Create(retries int, delay time.Duration) []time.Duration {
	if e.cache == nil {
		e.cache = retrier.ExponentialBackoff(retries, delay)
	}

	return e.cache
}
