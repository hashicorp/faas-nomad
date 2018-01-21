package consul

import (
	"log"

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
	m.data = args.Get(0).(chan []*dependency.CatalogService)

	for cs := range m.data {
		log.Println("dep:", args.Get(1).(dependency.Dependency).String())
		f(args.Get(1).(dependency.Dependency), cs)
	}
}
