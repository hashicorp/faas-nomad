package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
)

func getFunctions(
	client nomad.Allocations,
	allocs []*api.AllocationListStub) ([]requests.Function, error) {

	functions := make([]requests.Function, 0)
	for _, a := range allocs {

		if a.ClientStatus == "running" {
			allocation, _, err := client.Info(a.ID, nil)
			if err != nil {
				return functions, err
			}

			functions = append(functions, requests.Function{
				Name:            allocation.Job.TaskGroups[0].Tasks[0].Name,
				Image:           allocation.Job.TaskGroups[0].Tasks[0].Config["image"].(string),
				Replicas:        uint64(*allocation.Job.TaskGroups[0].Count),
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

// MakeReader implements the OpenFaaS reader handler
func MakeReader(client nomad.Allocations) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Not sure if prefix is the right option
		options := api.QueryOptions{}
		options.Prefix = "faas_function"

		allocations, _, err := client.List(nil)
		if err != nil {
			writeError(w, err)
			return
		}

		functions, err := getFunctions(client, allocations)
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
