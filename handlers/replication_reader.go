package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
)

// MakeReplicationReader creates a replication reader handler
func MakeReplicationReader(
	client nomad.Job,
	getVars func(*http.Request) map[string]string) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		vars := getVars(r)
		functionName := vars["name"]

		if functionName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, fmt.Errorf("No function name"))
			return
		}

		job, _, err := client.Info(nomad.JobPrefix+functionName, nil)
		if job == nil {
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
