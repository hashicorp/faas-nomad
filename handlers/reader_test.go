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

var mockDeployments *nomad.MockDeployments

func setupReader() (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockDeployments = &nomad.MockDeployments{}

	return MakeReader(mockDeployments),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte("")))
}

func TestHandlerReturns500OnClientError(t *testing.T) {
	handler, rw, r := setupReader()
	mockDeployments.On("List", mock.Anything).Return(make([]*api.Deployment, 0), nil, fmt.Errorf("BOOM"))

	handler(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestHandlerReturnsDeployments(t *testing.T) {
	handler, rw, r := setupReader()

	d := make([]*api.Deployment, 0)
	d = append(d, &api.Deployment{})
	mockDeployments.On("List", mock.Anything).Return(d, nil, fmt.Errorf("BOOM"))

	handler(rw, r)

	body, err := ioutil.ReadAll(rw.Body)
	if err != nil {
		t.Fatal(err)
	}

	funcs := make([]requests.Function, 0)
	json.Unmarshal(body, funcs)

	assert.True(t, false)
}
