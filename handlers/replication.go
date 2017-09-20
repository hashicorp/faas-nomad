package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexellis/faas-provider/types"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
)

// MakeReplicationReader creates a replication reader handler
func MakeReplicationReader(client nomad.Job) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {

		job, err := getJob(client, r)
		if job == nil || err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err)
			return
		}

		resp := requests.Function{}
		resp.Name = strings.Replace(*job.ID, nomad.JobPrefix, "", -1)
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(resp)
	}
}

// MakeReplicationWriter creates a handler for scaling functions
func MakeReplicationWriter(client nomad.Job) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {

		job, err := getJob(client, r)
		if job == nil || err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err)
			return
		}

		req := types.ScaleServiceRequest{}
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, err)
			return
		}

		// update nomad job
	}
}

func getJob(client nomad.Job, r *http.Request) (*api.Job, error) {
	functionName := r.Context().Value(FunctionNameCTXKey).(string)

	job, _, err := client.Info(nomad.JobPrefix+functionName, nil)
	if err != nil {
		return nil, err
	}

	return job, nil
}
