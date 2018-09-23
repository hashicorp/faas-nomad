package handlers

import (
	"encoding/json"

	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/openfaas/faas/gateway/requests"
)

var mockJob *nomad.MockJob
var mockServiceResolver *consul.MockResolver

type testFunctionRequest struct {
	requests.CreateFunctionRequest
}

func (r testFunctionRequest) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

func createRequest() testFunctionRequest {
	req := testFunctionRequest{}
	req.Service = "TestFunction"
	req.Labels = &map[string]string{}
	req.Constraints = []string{}
	return req
}

func deleteRequest() string {
	req := requests.DeleteFunctionRequest{}
	req.FunctionName = "TestFunction"

	data, _ := json.Marshal(req)
	return string(data)
}
