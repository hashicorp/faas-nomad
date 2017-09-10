package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/consul"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockProxyClient *MockProxyClient
var mockServiceResolver *consul.MockServiceResolver

func setupProxy(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockProxyClient = &MockProxyClient{}
	mockServiceResolver = &consul.MockServiceResolver{}

	r := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(body)))
	rr := httptest.NewRecorder()

	return MakeProxy(mockProxyClient, mockServiceResolver), rr, r
}

func TestProxyHandlerOnGETReturnsBadRequest(t *testing.T) {
	h, rr, r := setupProxy("")
	r.Method = "GET"

	h(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProxyHandlerWithNoFunctionNameReturnsBadRequest(t *testing.T) {
	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("")

	h(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProxyHandlerWithFunctionNameCallsResolve(t *testing.T) {
	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).Return([]byte{})
	mockServiceResolver.On("Resolve", "function").Return([]string{"testaddress"})

	h(rr, r)

	mockServiceResolver.AssertCalled(t, "Resolve", "function")
}

func TestProxyHandlerCallsCallAndRetrunResponse(t *testing.T) {
	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).Return([]byte{})
	mockServiceResolver.On("Resolve", "function").Return([]string{"testaddress"})

	h(rr, r)

	mockProxyClient.AssertCalled(t, "CallAndReturnResponse", "testaddress", rr, r)
}

func TestProxyHandlerWithLocalhostReplacesWithDockerMacAddress(t *testing.T) {
	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).Return([]byte{})
	mockServiceResolver.On("Resolve", "function").Return([]string{"127.0.0.1"})

	h(rr, r)

	mockProxyClient.AssertCalled(t, "CallAndReturnResponse", "docker.for.mac.localhost", rr, r)
}
