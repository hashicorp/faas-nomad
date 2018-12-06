package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeReplicationReader creates a replication reader handler
func MakeReplicationReader(client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	log := logger.Named("replicationreader_handler")

	return func(rw http.ResponseWriter, r *http.Request) {
		stats.Incr("replicationreader.called", nil, 1)

		job, err := getJob(client, r)
		if job == nil || err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err)

			log.Error("Error getting job", "error", err)
			stats.Incr("replicationreader.error.notfound", nil, 1)
			return
		}

		// get the number of available allocations from the job
		allocs, err := getAllocationReadyCount(client, job, r)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, err)

			log.Error("Error getting job allocations", "error", err)
			stats.Incr("replicationreader.error.internalerror", nil, 1)
			return
		}

		resp := requests.Function{
			Name:              sanitiseJobName(job),
			Image:             job.TaskGroups[0].Tasks[0].Config["image"].(string),
			Replicas:          uint64(*job.TaskGroups[0].Count),
			AvailableReplicas: allocs,
			Labels:            parseLabels(job.TaskGroups[0].Tasks[0].Config["labels"].([]interface{})),
			Annotations:       &job.Meta,
		}

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(resp)

		stats.Incr("replicationreader.success", nil, 1)
	}
}

// MakeReplicationWriter creates a handler for scaling functions
func MakeReplicationWriter(client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	log := logger.Named("replicationwriter_handler")

	return func(rw http.ResponseWriter, r *http.Request) {
		stats.Incr("replicationwriter.called", nil, 1)

		job, err := getJob(client, r)
		if job == nil || err != nil {
			rw.WriteHeader(http.StatusNotFound)

			log.Error("Error getting job", "error", err)
			stats.Incr("replicationwriter.error.notfound", nil, 1)
			return
		}

		req := types.ScaleServiceRequest{}
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.ServiceName == "" {
			rw.WriteHeader(http.StatusBadRequest)

			log.Error("Bad request", "error", err)
			stats.Incr("replicationwriter.error.badrequest", nil, 1)
			return
		}

		// update nomad job
		log.Info("Updating function", "function", req.ServiceName, "scale", req.Replicas)

		replicas := int(req.Replicas)
		job.TaskGroups[0].Count = &replicas

		_, _, err = client.Register(job, nil)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)

			log.Error("Error updating job", "error", err)
			stats.Incr("replicationwriter.error.internalerror", nil, 1)
		}

		stats.Gauge("deploy.count", float64(req.Replicas), []string{"job:" + req.ServiceName}, 1)
		stats.Incr("replicationwriter.success", nil, 1)
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

func getAllocationReadyCount(client nomad.Job, job *api.Job, r *http.Request) (uint64, error) {
	allocs, _, err := client.Allocations(*job.ID, true, nil)
	var readyCount uint64

	for _, a := range allocs {
		for _, ts := range a.TaskStates {
			if ts.State == "running" {
				readyCount += 1
			}
		}
	}

	return readyCount, err
}
