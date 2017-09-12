package ultraclient

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockBackoffStrategy is a mock implementation of the BackoffStrategy interface
type MockBackoffStrategy struct {
	mock.Mock
}

// Create is a mock implementation of Create
func (m *MockBackoffStrategy) Create(retries int, delay time.Duration) []time.Duration {
	args := m.Called(retries, delay)

	return args.Get(0).([]time.Duration)
}
