package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/nomad/api"
	"github.com/nicholasjackson/faas-nomad/nomad"
)

// MakeReader implements the OpenFaaS reader handler
func MakeReader(allocs nomad.Allocations) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request: " + r.URL.RawPath)

		// Not sure if prefix is the right option
		options := api.QueryOptions{}
		options.Prefix = "faas_function"

		allocations, _, err := allocs.List(&options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		functions := make([]requests.Function, 0)
		for _, a := range allocations {
			allocation, _, err := allocs.Info(a.ID, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			functions = append(functions, requests.Function{
				Name:            allocation.Name,
				Image:           allocation.Job.TaskGroups[0].Tasks[0].Config["image"].(string),
				Replicas:        uint64(*allocation.Job.TaskGroups[0].Count),
				InvocationCount: 0,
			})
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}
