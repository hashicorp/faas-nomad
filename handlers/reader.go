package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
)

// MakeReader implements the OpenFaaS reader handler
func MakeReader(client nomad.Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Not sure if prefix is the right option
		options := &api.QueryOptions{}
		options.Prefix = nomad.JobPrefix

		jobs, _, err := client.List(options)
		if err != nil {
			writeError(w, err)
			return
		}

		functions, err := getFunctions(client, jobs)
		if err != nil {
			writeError(w, err)
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}

func getFunctions(
	client nomad.Job,
	jobs []*api.JobListStub) ([]requests.Function, error) {

	functions := make([]requests.Function, 0)
	for _, j := range jobs {

		if j.Status == "running" {
			job, _, err := client.Info(j.ID, nil)
			if err != nil {
				return functions, err
			}

			jobName := strings.Replace(
				job.TaskGroups[0].Tasks[0].Name,
				nomad.JobPrefix,
				"",
				-1)

			functions = append(functions, requests.Function{
				Name:            jobName,
				Image:           job.TaskGroups[0].Tasks[0].Config["image"].(string),
				Replicas:        uint64(*job.TaskGroups[0].Count),
				InvocationCount: 0,
			})
		}

	}

	return functions, nil
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
	log.Println(err)
	return
}
