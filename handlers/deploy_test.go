package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicholasjackson/faas-nomad/nomad"
	"github.com/stretchr/testify/assert"
)

var mockJob *nomad.MockJob

func setupDeploy(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}

	return MakeDeploy(mockJob),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte(body)))
}

func TestHandlerReturnsErrorOnInvalidRequest(t *testing.T) {
	h, rw, r := setupDeploy("")

	h(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}
