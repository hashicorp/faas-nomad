package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupReader() (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	mockStatsD := &metrics.MockStatsD{}
	mockStatsD.On("Incr", mock.Anything, mock.Anything, mock.Anything)

	return MakeReader(mockJob, mockStatsD),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte("")))
}

func createMockJob(id string, count int) *api.Job {
	status := "running"
	name := nomad.JobPrefix + "JOB123"
	return &api.Job{
		ID:     &name,
		Status: &status,
		TaskGroups: []*api.TaskGroup{&api.TaskGroup{
			Count: &count,
			Tasks: []*api.Task{&api.Task{
				Name:   "Task" + id,
				Config: map[string]interface{}{"image": "docker"},
			}},
		},
		}}
}

func TestHandlerReturns500OnClientListError(t *testing.T) {
	handler, rw, r := setupReader()
	mockJob.On("List", mock.Anything).Return(make([]*api.JobListStub, 0), nil, fmt.Errorf("BOOM"))

	handler(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestHandlerReturns500OnClientInfoError(t *testing.T) {
	handler, rw, r := setupReader()

	a1 := createMockJob("1234", 1)

	d := make([]*api.JobListStub, 0)
	d = append(d, &api.JobListStub{ID: *a1.ID, Status: *a1.Status})

	mockJob.On("List", mock.Anything).Return(d, nil, nil)
	mockJob.On("Info", *a1.ID, mock.Anything).Return(nil, nil, fmt.Errorf("BOOM"))

	handler(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestHandlerReturnsDeployments(t *testing.T) {
	handler, rw, r := setupReader()

	a1 := createMockJob("1234", 1)
	d := make([]*api.JobListStub, 0)
	d = append(d, &api.JobListStub{ID: *a1.ID, Status: *a1.Status})

	mockJob.On("List", mock.Anything).Return(d, nil, nil)
	mockJob.On("Info", *a1.ID, mock.Anything).Return(a1, nil, nil)

	handler(rw, r)

	body, err := ioutil.ReadAll(rw.Body)
	if err != nil {
		t.Fatal(err)
	}

	funcs := make([]requests.Function, 0)
	json.Unmarshal(body, &funcs)

	jobName := strings.Replace(a1.TaskGroups[0].Tasks[0].Name, nomad.JobPrefix, "", -1)
	assert.Equal(t, jobName, funcs[0].Name)
	assert.Equal(t, a1.TaskGroups[0].Tasks[0].Config["image"].(string), funcs[0].Image)
	assert.Equal(t, uint64(*a1.TaskGroups[0].Count), funcs[0].Replicas)
}

func TestHandlerReturnsRunningDeployments(t *testing.T) {
	handler, rw, r := setupReader()

	a1 := createMockJob("1234", 1)
	a2 := createMockJob("4567", 1)
	a3 := createMockJob("8929", 1)

	a2status := "stopped"
	a2.Status = &a2status

	a3status := "pending"
	a3.Status = &a3status

	d := make([]*api.JobListStub, 0)
	d = append(d, &api.JobListStub{ID: *a1.ID, Status: *a1.Status})
	d = append(d, &api.JobListStub{ID: *a2.ID, Status: *a2.Status})
	d = append(d, &api.JobListStub{ID: *a3.ID, Status: *a3.Status})

	mockJob.On("List", mock.Anything).Return(d, nil, nil)
	mockJob.On("Info", *a1.ID, mock.Anything).Return(a1, nil, nil)
	mockJob.On("Info", *a2.ID, mock.Anything).Return(a2, nil, nil)

	handler(rw, r)

	body, err := ioutil.ReadAll(rw.Body)
	if err != nil {
		t.Fatal(err)
	}

	funcs := make([]requests.Function, 0)
	json.Unmarshal(body, &funcs)

	assert.Equal(t, 2, len(funcs))
}
