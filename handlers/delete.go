package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/openfaas/faas/gateway/requests"
)

const functionNamespace string = "default"

// MakeDelete creates a handler for deploying functions
func MakeDelete(sr consul.ServiceResolver, client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	log := logger.Named("delete_handler")

	return func(w http.ResponseWriter, r *http.Request) {
		stats.Incr("delete.called", nil, 1)

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		req := requests.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &req)
		if err != nil || req.FunctionName == "" {
			w.WriteHeader(http.StatusBadRequest)

			log.Error("Bad request", "error", err)
			stats.Incr("delete.error.badrequest", []string{"job:" + req.FunctionName}, 1)
			return
		}

		log.Info("Deleting function", "function", req.FunctionName)

		// Delete job /v1/jobs
		_, _, err = client.Deregister(nomad.JobPrefix+req.FunctionName, false, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))

			log.Error("Error deregistering function", "error", err)
			stats.Incr("delete.error.deregister", []string{"job:" + req.FunctionName}, 1)
			return
		}

		sr.RemoveCacheItem(req.FunctionName)

		stats.Gauge("deploy.count", 0, []string{"job:" + req.FunctionName}, 1)
		stats.Incr("delete.success", []string{"job:" + req.FunctionName}, 1)
	}
}
