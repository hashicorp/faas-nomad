package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDelete(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	mockStats := &metrics.MockStatsD{}
	mockStats.On("Incr", mock.Anything, mock.Anything, mock.Anything)
	mockStats.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	mockServiceResolver = &consul.MockResolver{}
	mockServiceResolver.On("RemoveCacheItem", mock.Anything)

	logger := hclog.Default()

	return MakeDelete(mockServiceResolver, mockJob, logger, mockStats),
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
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestDeleteHandlerReturnsErrorOnAPIError(t *testing.T) {
	h, rw, r := setupDelete(deleteRequest())
	mockJob.On("Deregister", nomad.JobPrefix+"TestFunction", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("BOOM"))

	h(rw, r)

	mockJob.AssertCalled(t, "Deregister", nomad.JobPrefix+"TestFunction", mock.Anything, mock.Anything)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestDeleteHandlerClearsConsulCache(t *testing.T) {
	h, rw, r := setupDelete(deleteRequest())
	mockJob.On("Deregister", nomad.JobPrefix+"TestFunction", mock.Anything, mock.Anything).Return(nil, nil, nil)

	h(rw, r)

	mockServiceResolver.AssertCalled(t, "RemoveCacheItem", "TestFunction")
	assert.Equal(t, http.StatusOK, rw.Code)
}
