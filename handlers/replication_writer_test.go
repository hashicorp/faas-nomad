package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexellis/faas-provider/types"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupReplicationWriter(functionName string) (
	http.HandlerFunc,
	*httptest.ResponseRecorder,
	*http.Request) {

	mockJob = &nomad.MockJob{}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test/test_function", nil)
	r = r.WithContext(context.WithValue(r.Context(), FunctionNameCTXKey, functionName))

	h := MakeReplicationWriter(mockJob)

	return h, rr, r
}

func TestReplicationWReturnsNotFoundWhenNoFunction(t *testing.T) {
	h, rr, r := setupReplicationWriter("")
	mockJob.On("Info", mock.Anything, mock.Anything).Return(nil, nil, nil)

	h(rr, r)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestReplicationWReturnsBadRequestWhenNoBody(t *testing.T) {
	h, rr, r := setupReplicationWriter("testFunc")
	mockJob.On("Info", mock.Anything, mock.Anything).Return(&api.Job{}, nil, nil)

	h(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestReplicationWUpdatesNomadJob(t *testing.T) {
	count := 1
	job := &api.Job{
		TaskGroups: []*api.TaskGroup{
			&api.TaskGroup{Count: &count},
		},
	}

	h, rr, r := setupReplicationWriter("testFunc")
	mockJob.On("Info", mock.Anything, mock.Anything).Return(job, nil, nil)

	req := types.ScaleServiceRequest{Replicas: 2}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	//	r.Body = data
	h(rr, r)

	count++
	job.TaskGroups[0].Count = &count
	mockJob.AssertCalled(t, "Register", job, mock.Anything)
}
