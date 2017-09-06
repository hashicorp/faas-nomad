package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/nomad/api"
	"github.com/nicholasjackson/faas-nomad/nomad"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockAllocations *nomad.MockAllocations

func setupReader() (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockAllocations = &nomad.MockAllocations{}

	return MakeReader(mockAllocations),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte("")))
}

func TestHandlerReturns500OnClientListError(t *testing.T) {
	handler, rw, r := setupReader()
	mockAllocations.On("List", mock.Anything).Return(make([]*api.AllocationListStub, 0), nil, fmt.Errorf("BOOM"))

	handler(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestHandlerReturns500OnClientInfoError(t *testing.T) {
	handler, rw, r := setupReader()

	a1 := &api.Allocation{ID: "1234", Name: "Test1234"}

	d := make([]*api.AllocationListStub, 0)
	d = append(d, &api.AllocationListStub{ID: a1.ID})

	mockAllocations.On("List", mock.Anything).Return(d, nil, nil)
	mockAllocations.On("Info", a1.ID, mock.Anything).Return(nil, nil, fmt.Errorf("BOOM"))

	handler(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestHandlerReturnsDeployments(t *testing.T) {
	handler, rw, r := setupReader()

	a1Count := 1
	a1 := &api.Allocation{
		ID:   "1234",
		Name: "Test1234",
		Job: &api.Job{TaskGroups: []*api.TaskGroup{&api.TaskGroup{
			Count: &a1Count,
			Tasks: []*api.Task{&api.Task{Config: map[string]interface{}{"image": "docker"}}},
		}}},
	}

	d := make([]*api.AllocationListStub, 0)
	d = append(d, &api.AllocationListStub{ID: a1.ID})

	mockAllocations.On("List", mock.Anything).Return(d, nil, nil)
	mockAllocations.On("Info", a1.ID, mock.Anything).Return(a1, nil, nil)

	handler(rw, r)

	body, err := ioutil.ReadAll(rw.Body)
	if err != nil {
		t.Fatal(err)
	}

	funcs := make([]requests.Function, 0)
	json.Unmarshal(body, &funcs)

	assert.Equal(t, a1.Name, funcs[0].Name)
	assert.Equal(t, a1.Job.TaskGroups[0].Tasks[0].Config["image"].(string), funcs[0].Image)
	assert.Equal(t, uint64(*a1.Job.TaskGroups[0].Count), funcs[0].Replicas)
}
