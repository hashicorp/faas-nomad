package nomad

import (
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/mock"
)

// MockAllocations is a mock implementation of the Nomad allocations api
type MockAllocations struct {
	mock.Mock
}

// List returns a list of allocations
func (m *MockAllocations) List(q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error) {
	args := m.Called(q)

	var md *api.QueryMeta
	if m := args.Get(1); m != nil {
		md = m.(*api.QueryMeta)
	}

	return args.Get(0).([]*api.AllocationListStub), md, args.Error(2)
}

// Info returns a single allocation from the API
func (m *MockAllocations) Info(allocID string, q *api.QueryOptions) (*api.Allocation, *api.QueryMeta, error) {
	args := m.Called(allocID, q)

	var md *api.QueryMeta
	if m := args.Get(1); m != nil {
		md = m.(*api.QueryMeta)
	}

	var alloc *api.Allocation
	if a := args.Get(0); a != nil {
		alloc = a.(*api.Allocation)
	}

	return alloc, md, args.Error(2)
}
