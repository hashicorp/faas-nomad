package metrics

import "github.com/stretchr/testify/mock"

// MockStatsD is a Mock which implements the StatsD interface
type MockStatsD struct {
	mock.Mock
}

// Incr calls the mock method to increment a statistic
func (m *MockStatsD) Incr(name string, tags []string, rate float64) error {
	m.Mock.Called(name, tags, rate)

	return nil
}
