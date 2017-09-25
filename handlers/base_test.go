package handlers

import (
	"encoding/json"

	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/openfaas/faas/gateway/requests"
)

var mockJob *nomad.MockJob

func createRequest() string {
	req := requests.CreateFunctionRequest{}
	req.Service = "TestFunction"

	data, _ := json.Marshal(req)
	return string(data)
}

func deleteRequest() string {
	req := requests.DeleteFunctionRequest{}
	req.FunctionName = "TestFunction"

	data, _ := json.Marshal(req)
	return string(data)
}
