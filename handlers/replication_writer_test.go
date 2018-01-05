package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupReplicationWriter(t *testing.T, functionName string, req *types.ScaleServiceRequest) (
	http.HandlerFunc,
	*httptest.ResponseRecorder,
	*http.Request) {

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	mockJob = &nomad.MockJob{}
	mockStats := &metrics.MockStatsD{}
	mockStats.On("Incr", mock.Anything, mock.Anything, mock.Anything)
	mockStats.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test/test_function", bytes.NewReader(body))
	r = r.WithContext(context.WithValue(r.Context(), FunctionNameCTXKey, functionName))

	logger := hclog.Default()

	h := MakeReplicationWriter(mockJob, logger, mockStats)

	return h, rr, r
}

func TestReplicationWReturnsNotFoundWhenNoFunction(t *testing.T) {
	h, rr, r := setupReplicationWriter(t, "", nil)
	mockJob.On("Info", mock.Anything, mock.Anything).Return(nil, nil, nil)

	h(rr, r)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestReplicationWReturnsBadRequestWhenNoBody(t *testing.T) {
	h, rr, r := setupReplicationWriter(t, "testFunc", nil)
	mockJob.On("Info", mock.Anything, mock.Anything).Return(&api.Job{}, nil, nil)

	h(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestReplicationWUpdatesNomadJob(t *testing.T) {
	count := 1
	job := api.Job{
		TaskGroups: []*api.TaskGroup{
			&api.TaskGroup{Count: &count},
		},
	}

	req := types.ScaleServiceRequest{Replicas: 2, ServiceName: "testFunc"}
	h, rr, r := setupReplicationWriter(t, "testFunc", &req)

	mockJob.On("Info", mock.Anything, mock.Anything).Return(&job, nil, nil)
	mockJob.On("Register", mock.Anything, mock.Anything).Return(nil, nil, nil)

	h(rr, r)

	args := mockJob.Calls[1].Arguments
	j := args.Get(0).(*api.Job)

	mockJob.AssertCalled(t, "Register", &job, mock.Anything)
	assert.Equal(t, 2, *j.TaskGroups[0].Count)
}

func TestReplicationWReturnsInternalServerErrorOnRegisterError(t *testing.T) {
	count := 1
	job := api.Job{
		TaskGroups: []*api.TaskGroup{
			&api.TaskGroup{Count: &count},
		},
	}

	req := types.ScaleServiceRequest{Replicas: 2, ServiceName: "testFunc"}
	h, rr, r := setupReplicationWriter(t, "testFunc", &req)

	mockJob.On("Info", mock.Anything, mock.Anything).Return(&job, nil, nil)
	mockJob.On("Register", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("BOOM"))

	h(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
