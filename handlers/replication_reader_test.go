package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupReplicationReader(functionName string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	mockStats := &metrics.MockStatsD{}
	mockStats.On("Incr", mock.Anything, mock.Anything, mock.Anything)

	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test/test_function", nil)
	r = r.WithContext(context.WithValue(r.Context(), FunctionNameCTXKey, functionName))

	logger := hclog.Default()

	h := MakeReplicationReader(mockJob, logger, mockStats)

	return h, rr, r
}

func TestMiddlewareReturns404WhenNotFound(t *testing.T) {
	functionName := "notFound"
	h, rr, r := setupReplicationReader(functionName)
	mockJob.On("Info", nomad.JobPrefix+functionName, mock.Anything).Return(nil, nil, nil)

	h(rr, r)

	mockJob.AssertCalled(t, "Info", nomad.JobPrefix+functionName, mock.Anything)
	assert.Equal(t, rr.Code, http.StatusNotFound)
}

func TestReplicationReturnsFunctionWhenFound(t *testing.T) {
	functionName := "tester"
	jobName := nomad.JobPrefix + functionName

	h, rr, r := setupReplicationReader(functionName)
	mockJob.On("Info", jobName, mock.Anything).
		Return(&api.Job{ID: &jobName}, nil, nil)

	h(rr, r)

	f := &requests.Function{}
	err := json.NewDecoder(rr.Body).Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, functionName, f.Name)
}
