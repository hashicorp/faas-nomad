package ultraclient

import (
	"net/url"

	"github.com/stretchr/testify/mock"
)

// MockClient implements a mock implementation of the ultraclient
type MockClient struct {
	mock.Mock
}

// Do is the mock execution of the Do method
// mockClient.On("Do").Return(error, url)
func (m *MockClient) Do(work WorkFunc) error {
	args := m.Called(work)

	if len(args) > 1 {
		return work(args.Get(1).(url.URL))
	}

	return args.Error(0)
}

// UpdateEndpoints is a mock execution of the interface method
func (m *MockClient) UpdateEndpoints(endpoints []url.URL) {
	m.Called(endpoints)
}

// Clone is the mock execution of the Clone method, returns self
func (m *MockClient) Clone() Client {
	m.Called()
	return m
}

// RegisterStats is the mock execution of the RegisterStats method
func (m *MockClient) RegisterStats(stats Stats) {
	m.Called(stats)
}
