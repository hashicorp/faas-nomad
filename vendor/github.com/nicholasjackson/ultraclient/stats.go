package ultraclient

import (
	"time"

	"github.com/stretchr/testify/mock"
)

const (
	//StatsTiming is a statsD tag for timing metrics
	StatsTiming = "timing"
	// StatsCalled is a statsD tag to indicate a method has been called
	StatsCalled = "called"
	// StatsSuccess is a statsD tag to indicate operational success
	StatsSuccess = "success"
	// StatsRetry is a statsD tag to indicate the client will retry
	StatsRetry = "retry"
	// StatsError is a statsD tag to indicate that the client has resulted in
	// an error
	StatsError = "error"
	// StatsCircuitOpen is a statsD tag to indicate that the client has an open
	// circuit
	StatsCircuitOpen = "circuitopen"
	// StatsTimeout is a statsD tag to indicate that the operation has timed out
	StatsTimeout = "timeout"
)

// Stats is an interface which the concrete type will implement in order to send statistics to
// endpoints like StatsD or Logging
type Stats interface {
	// Increment is a simple incremental counter for the given bucket
	// name is the name of the bucket to write to
	// tags is the list of tags to associate with the metric
	// rate is the rate to associate with the metric
	Increment(name string, tags []string, rate float64)

	// Timing records the duration of the given function
	// name is the name of the bucket to write to
	// tags is the list of tags to associate with the metric
	// duration is the duration for the call
	// rate is the rate to associate with the metric
	Timing(name string, tags []string, duration time.Duration, rate float64)
}

// MockStats is a mock implementation of the Stats interface to be used
// for testing
type MockStats struct {
	mock.Mock
}

// Increment is a mock implementation of the Stats interface
func (m *MockStats) Increment(name string, tags []string, rate float64) {
	m.Called(name, tags, rate)
}

// Timing is a mock implementation of the Stats interface
func (m *MockStats) Timing(name string, tags []string, duration time.Duration, rate float64) {
	m.Called(name, tags, duration, rate)
}
