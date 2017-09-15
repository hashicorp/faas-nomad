package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var functionName = ""

func setupReplicationReader() (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test/test_function", nil)

	h := MakeReplicationReader(mockJob, func(*http.Request) map[string]string {
		return map[string]string{"name": functionName}
	})

	return h, rr, r
}

func TestReplicationRReturnsBadRequestWhenNoFunction(t *testing.T) {
	h, rr, r := setupReplicationReader()

	h(rr, r)

	assert.Equal(t, rr.Code, http.StatusBadRequest)
}

func TestReplicationRReturns404WhenNotFound(t *testing.T) {
	h, rr, r := setupReplicationReader()
	functionName = "notFound"
	mockJob.On("Info", nomad.JobPrefix+functionName, mock.Anything).Return(nil, nil, nil)

	h(rr, r)

	mockJob.AssertCalled(t, "Info", nomad.JobPrefix+functionName, mock.Anything)
	assert.Equal(t, rr.Code, http.StatusNotFound)
}

func TestReplicationRReturnsFunctionWhenFound(t *testing.T) {
	h, rr, r := setupReplicationReader()
	functionName = "found"
	mockJob.On("Info", nomad.JobPrefix+functionName, mock.Anything).Return(&api.Job{ID: &functionName}, nil, nil)

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
