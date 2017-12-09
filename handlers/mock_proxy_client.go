package handlers

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockProxyClient is a mock implementation of the ProxyClient
type MockProxyClient struct {
	mock.Mock
}

func (mp *MockProxyClient) GetFunctionName(r *http.Request) string {
	args := mp.Called(r)

	return args.Get(0).(string)
}

// CallAndReturnResponse returns a mock respoonse
func (mp *MockProxyClient) CallAndReturnResponse(address string, body []byte, h http.Header) (
	[]byte, http.Header, error) {
	args := mp.Called(address, body, h)

	return args.Get(0).([]byte), args.Get(1).(http.Header), args.Error(2)
}
