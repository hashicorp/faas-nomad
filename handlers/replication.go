package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas-provider/types"
	"github.com/openfaas/faas/gateway/requests"
)

// MakeReplicationReader creates a replication reader handler
func MakeReplicationReader(client nomad.Job, stats metrics.StatsD) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		stats.Incr("replicationreader.called", nil, 1)

		job, err := getJob(client, r)
		if job == nil || err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err)

			stats.Incr("replicationreader.error.notfound", nil, 1)
			return
		}

		resp := requests.Function{}
		resp.Name = strings.Replace(*job.ID, nomad.JobPrefix, "", -1)
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(resp)

		stats.Incr("replicationreader.success", nil, 1)
	}
}

// MakeReplicationWriter creates a handler for scaling functions
func MakeReplicationWriter(client nomad.Job, stats metrics.StatsD) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		stats.Incr("replicationwriter.called", nil, 1)

		job, err := getJob(client, r)
		if job == nil || err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err)

			stats.Incr("replicationwriter.error.notfound", nil, 1)
			return
		}

		req := types.ScaleServiceRequest{}
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.ServiceName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, err)

			stats.Incr("replicationwriter.error.badrequest", nil, 1)
			return
		}

		// update nomad job
		log.Println("Updating function to scale:", req.Replicas)
		replicas := int(req.Replicas)
		job.TaskGroups[0].Count = &replicas

		_, _, err = client.Register(job, nil)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, err)

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
