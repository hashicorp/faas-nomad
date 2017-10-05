package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
	r = r.WithContext(context.WithValue(r.Context(), FunctionNameCTXKey, "function"))
	rr := httptest.NewRecorder()

	return MakeProxy(mockProxyClient, mockServiceResolver, "", nil), rr, r
}

func TestProxyHandlerOnGETReturnsBadRequest(t *testing.T) {
	h, rr, r := setupProxy("")
	r.Method = "GET"

	h(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProxyHandlerWithFunctionNameCallsResolve(t *testing.T) {
	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte{}, http.Header{}, nil)
	mockServiceResolver.On("Resolve", "function").Return([]string{"http://testaddress"})

	h(rr, r)

	mockServiceResolver.AssertCalled(t, "Resolve", "function")
}

func TestProxyHandlerCallsCallAndReturnResponse(t *testing.T) {
	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte{}, http.Header{}, nil)
	mockServiceResolver.On("Resolve", "function").Return([]string{"http://testaddress"})

	h(rr, r)

	mockProxyClient.AssertCalled(t, "CallAndReturnResponse", "http://testaddress", mock.Anything, mock.Anything)
}

func TestProxyHandlerReturnsErrorWhenNoEndpoints(t *testing.T) {
	f := setEnv()
	defer f()

	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte{}, http.Header{}, nil)
	mockServiceResolver.On("Resolve", "function").Return([]string{})

	h(rr, r)

	mockProxyClient.AssertNotCalled(t, "CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestProxyHandlerReturnsErrorWhenClientError(t *testing.T) {
	f := setEnv()
	defer f()

	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte{}, http.Header{}, fmt.Errorf("Oops, I did it again"))
	mockServiceResolver.On("Resolve", "function").Return([]string{"http://testaddress"})

	h(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestProxyHandlerSetsHeadersOnSuccess(t *testing.T) {
	f := setEnv()
	defer f()

	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte{}, http.Header{"TestHeader": []string{"Testiculous"}}, nil)
	mockServiceResolver.On("Resolve", "function").Return([]string{"http://testaddress"})

	h(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Testiculous", rr.Header().Get("TestHeader"))
}

func TestProxyHandlerSetsBodyOnSuccess(t *testing.T) {
	f := setEnv()
	defer f()

	h, rr, r := setupProxy("")
	mockProxyClient.On("GetFunctionName", mock.Anything).Return("function")
	mockProxyClient.On("CallAndReturnResponse", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte("Something Something"), http.Header{}, nil)
	mockServiceResolver.On("Resolve", "function").Return([]string{"http://testaddress"})

	h(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Something Something", rr.Body.String())
}

func setEnv() func() {
	env := os.Environ()
	os.Setenv("HOST_IP", "myhost")

	return func() {
		for _, e := range env {
			kv := strings.Split(e, "=")
			os.Setenv(kv[0], kv[1])
		}
	}
}
