package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/nicholasjackson/faas-nomad/nomad"
)

// MakeDeploy creates a handler for deploying functions
func MakeDeploy(client nomad.Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create job /v1/jobs
	}
}
