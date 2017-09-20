package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var functionName = ""
var nextCalled = false
var contextValue interface{}

func setupFunctionMiddleware() (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	h := MakeExtractFunctionMiddleWare(func(*http.Request) map[string]string {
		vars := make(map[string]string)
		vars["name"] = functionName
		return vars
	},
		func(rw http.ResponseWriter, r *http.Request) {
			contextValue = r.Context().Value(FunctionNameCTXKey)
			nextCalled = true
		})

	rw := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	return h, rw, r
}

func TestMiddlewareReturnsBadRequestWhenNoFunction(t *testing.T) {
	h, rr, r := setupFunctionMiddleware()

	h(rr, r)

	assert.Equal(t, rr.Code, http.StatusBadRequest)
}

func TestMiddlewareSetsContext(t *testing.T) {
	functionName = "myfunction"
	h, rr, r := setupFunctionMiddleware()

	h(rr, r)

	assert.Equal(t, functionName, contextValue)
}
