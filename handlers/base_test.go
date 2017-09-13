package handlers

import (
	"encoding/json"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
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
