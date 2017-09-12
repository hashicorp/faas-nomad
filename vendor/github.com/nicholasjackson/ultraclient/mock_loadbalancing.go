package ultraclient

import (
	"net/url"
	"reflect"

	"github.com/stretchr/testify/mock"
)

// GetEndpoint is a function that can be passed as a return parameter to the
// On call for NextEndpoint, this can be used to control which url
// is returned in your tests
type GetEndpoint func() url.URL

// MockLoadbalancingStrategy is a mocked implementation of the Strategy interface for testing
// Usage:
// mock := MockLoadbalancingStrategy{}
// mock.On("NextEndpoint").Returns([]url.URL{url.URL{Host: ""}})
// mock.On("SetEndpoints", mock.Anything)
// mock.AssertCalled(t, "NextEndpoint")
type MockLoadbalancingStrategy struct {
	mock.Mock
}

// NextEndpoint returns the next endpoint in the list
func (m *MockLoadbalancingStrategy) NextEndpoint() url.URL {
	args := m.Called()

	arg := args.Get(0)
	if reflect.TypeOf(arg).Kind() == reflect.Func {
		return arg.(GetEndpoint)()
	}

	return args.Get(0).(url.URL)
}

// SetEndpoints sets the mocks internal register with the given arguments,
// this method can not be used to update the return values in NextEndpoint.
func (m *MockLoadbalancingStrategy) SetEndpoints(urls []url.URL) {
	m.Called(urls)
}

// Length returns the number of endpoints
func (m *MockLoadbalancingStrategy) Length() int {
	args := m.Called()
	return args.Get(0).(int)
}

// GetEndpoints returns the current endpoint collection
func (m *MockLoadbalancingStrategy) GetEndpoints() []url.URL {
	args := m.Called()
	return args.Get(0).([]url.URL)
}

// Clone creates a clone of the current object
func (m *MockLoadbalancingStrategy) Clone() LoadbalancingStrategy {
	m.Called()
	return &MockLoadbalancingStrategy{}
}
