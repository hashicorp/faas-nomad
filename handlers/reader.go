package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/nicholasjackson/faas-nomad/nomad"
)

// MakeReader implements the OpenFaaS reader handler
func MakeReader(client nomad.Deployments) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request: " + r.URL.RawPath)

		deployments, _, err := client.List(nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		functions := make([]requests.Function, 0)
		for _, d := range deployments {
			functions = append(functions, requests.Function{
				Name: d.JobID,
			})
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}
