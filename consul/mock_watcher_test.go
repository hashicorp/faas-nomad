package consul

import (
	"github.com/hashicorp/consul-template/dependency"
	"github.com/stretchr/testify/mock"
)

// MockWatcher implements the Watcher interface and is used to mock out consul template
type MockWatcher struct {
	mock.Mock
	data chan []*dependency.CatalogService
}

// Add adds a new dependency to the watch list
func (m *MockWatcher) Add(d dependency.Dependency) (bool, error) {
	args := m.Mock.Called(d)

	return args.Get(0).(bool), args.Error(1)
}

// Remove removes a dependency from the watch list
func (m *MockWatcher) Remove(d dependency.Dependency) bool {
	args := m.Mock.Called(d)

	return args.Get(0).(bool)
}

// ItterateDataCh itterates over the watchers data channel and calls a function
func (m *MockWatcher) ItterateDataCh(f itterateFunc) {
	args := m.Mock.Called(f)
	dep := args.Get(0).(dependency.Dependency)

	for cs := range m.data {
		f(dep, cs)
	}
}
