package ultraclient

import (
	"fmt"
	"net/url"

	"github.com/stretchr/testify/mock"
)

// MockCommand is a mock implementation of the Command interface
// This can be used for testing.
type MockCommand struct {
	mock.Mock
}

// Do calls the internal mock and returns the values specified in the setup
func (h *MockCommand) Do(url url.URL, input interface{}) (interface{}, error) {
	args := h.Called(url, input)
	fmt.Println(args)

	return args.Get(0), args.Error(1)
}
