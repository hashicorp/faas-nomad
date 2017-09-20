package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDelete(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}

	return MakeDelete(mockJob),
		httptest.NewRecorder(),
		httptest.NewRequest("DELETE", "/system/functions", bytes.NewReader([]byte(body)))
}

func TestDeleteHandlerReturnsErrorOnInvalidRequest(t *testing.T) {
	h, rw, r := setupDelete("")

	h(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestDeleteHandlerDeletesJob(t *testing.T) {
	h, rw, r := setupDelete(deleteRequest())
	mockJob.On("Deregister", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, nil)

	h(rw, r)

	mockJob.AssertCalled(t, "Deregister", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteHandlerReturnsErrorOnAPIError(t *testing.T) {
	h, rw, r := setupDelete(deleteRequest())
	mockJob.On("Deregister", nomad.JobPrefix+"TestFunction", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("BOOM"))

	h(rw, r)

	mockJob.AssertCalled(t, "Deregister", nomad.JobPrefix+"TestFunction", mock.Anything, mock.Anything)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
}
