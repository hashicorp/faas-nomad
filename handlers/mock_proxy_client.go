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

func (mp *MockProxyClient) CallAndReturnResponse(address string, rw http.ResponseWriter, r *http.Request) error {
	args := mp.Called(address, rw, r)

	if body := args.Get(0); body != nil {
		rw.Write(body.([]byte))
	}

	return args.Error(1)
}
