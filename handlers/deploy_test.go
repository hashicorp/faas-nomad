package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDeploy(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	mockJob.On("Register", mock.Anything, mock.Anything).Return(nil, nil, nil)

	return MakeDeploy(mockJob),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte(body)))
}

func TestHandlerReturnsErrorOnInvalidRequest(t *testing.T) {
	h, rw, r := setupDeploy("")

	h(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestHandlerRegistersJob(t *testing.T) {
	h, rw, r := setupDeploy(createRequest())

	h(rw, r)

	mockJob.AssertCalled(t, "Register", mock.Anything, mock.Anything)
}
