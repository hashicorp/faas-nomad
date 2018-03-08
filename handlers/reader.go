package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeReader implements the OpenFaaS reader handler
func MakeReader(client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	log := logger.Named("reader_handler")

	return func(w http.ResponseWriter, r *http.Request) {
		// Not sure if prefix is the right option
		options := &api.QueryOptions{}
		options.Prefix = nomad.JobPrefix

		log.Info("List functions called")
		stats.Incr("reader.called", nil, 1)

		jobs, _, err := client.List(options)
		if err != nil {
			writeError(w, err)

			log.Error("Error getting functions", "error", err)
			stats.Incr("reader.error.getjobs", nil, 1)
			return
		}

		functions, err := getFunctions(client, jobs)
		if err != nil {
			writeError(w, err)

			log.Error("Error getting functions", "error", err.Error())
			stats.Incr("reader.error.getfunctions", nil, 1)
			return
		}

		writeFunctionResponse(w, functions)

		log.Info("List functions success")
		stats.Incr("reader.success", nil, 1)
	}
}

func getFunctions(client nomad.Job, jobs []*api.JobListStub) ([]requests.Function, error) {
	functions := make([]requests.Function, 0)
	for _, j := range jobs {

		if j.Status == "running" || j.Status == "pending" {
			job, _, err := client.Info(j.ID, nil)
			if err != nil {
				return functions, err
			}

			functions = append(functions, requests.Function{
				Name:            sanitiseJobName(job),
				Image:           job.TaskGroups[0].Tasks[0].Config["image"].(string),
				Replicas:        uint64(*job.TaskGroups[0].Count),
				InvocationCount: 0,
			})
		}
	}

	return functions, nil
}

func sanitiseJobName(job *api.Job) string {
	return strings.Replace(job.TaskGroups[0].Tasks[0].Name, nomad.JobPrefix, "", -1)
}

func writeFunctionResponse(w http.ResponseWriter, fs []requests.Function) {
	functionBytes, _ := json.Marshal(fs)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(functionBytes)
}
func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
	log.Println(err)
	return
}
