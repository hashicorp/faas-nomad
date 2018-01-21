package consul

import (
	"fmt"

	"github.com/hashicorp/consul-template/dependency"
	"github.com/stretchr/testify/mock"
)

// MockServiceQuery is a mock implementation of the ConsulTemplate ServiceQuery interface
type MockServiceQuery struct {
	mock.Mock
	name string
}

// Fetch mocks the fetch interface on the consul service query
func (d *MockServiceQuery) Fetch(clients *dependency.ClientSet, opts *dependency.QueryOptions) (interface{}, *dependency.ResponseMetadata, error) {
	args := d.Mock.Called(clients)

	if args.Get(0) != nil {
		return args.Get(0).([]*dependency.CatalogService), nil, nil
	}

	return nil, nil, args.Error(2)
}

// CanShare implements the interface method CanShare
func (d *MockServiceQuery) CanShare() bool {
	return true
}

// Stop implements the interface method Stop
func (d *MockServiceQuery) Stop() {}

// String implements the interface method String
func (d *MockServiceQuery) String() string {
	return fmt.Sprintf("catalog.service(%s)", d.name)
}

// Type implements the interface method Type
func (d *MockServiceQuery) Type() dependency.Type {
	return dependency.TypeConsul
}
