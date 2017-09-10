package consul

import "github.com/stretchr/testify/mock"

// MockServiceResolver is a mock implementation of the ServiceResolver interface
type MockServiceResolver struct {
	mock.Mock
}

// Resolve returns the arguments from the Mocks setup method
func (mr *MockServiceResolver) Resolve(function string) ([]string, error) {
	args := mr.Called(function)

	if a := args.Get(0); a != nil {
		return a.([]string), nil
	}

	return nil, args.Error(1)
}
