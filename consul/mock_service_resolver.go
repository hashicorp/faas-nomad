package consul

import "github.com/stretchr/testify/mock"

// MockResolver is a mock implementation of the ServiceResolver interface
type MockResolver struct {
	mock.Mock
}

// Resolve returns the arguments from the Mocks setup method
func (mr *MockResolver) Resolve(function string) ([]string, error) {
	args := mr.Called(function)

	if a := args.Get(0); a != nil {
		return a.([]string), nil
	}

	return nil, args.Error(1)
}

// RemoveCacheItem implements the interface method for removing cache items
func (mr *MockResolver) RemoveCacheItem(function string) {
	mr.Called(function)
}
