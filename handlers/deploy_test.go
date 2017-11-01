package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDeploy(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	mockJob.On("Register", mock.Anything, mock.Anything).Return(nil, nil, nil)
	mockStats := &metrics.MockStatsD{}
	mockStats.On("Incr", mock.Anything, mock.Anything, mock.Anything)

	return MakeDeploy(mockJob, mockStats),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte(body)))
}

func TestHandlerReturnsErrorOnInvalidRequest(t *testing.T) {
	h, rw, r := setupDeploy("")

	h(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestHandlerRegistersJob(t *testing.T) {
	h, rw, r := setupDeploy(createRequest().String())

	h(rw, r)

	mockJob.AssertCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestHandlerRegistersWithEnvironmentVariables(t *testing.T) {
	reqEnv := map[string]string{"VAR1": "ABC"}
	fr := createRequest()
	fr.EnvVars = reqEnv

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	jobEnv := job.TaskGroups[0].Tasks[0].Env

	assert.Equal(t, reqEnv, jobEnv)
}

func TestHandlerRegistersWithFunctionProcess(t *testing.T) {
	fr := createRequest()
	fr.EnvProcess = "env"

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	jobEnv := job.TaskGroups[0].Tasks[0].Env

	assert.Equal(t, "env", jobEnv["fprocess"])
}
