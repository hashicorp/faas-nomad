package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
)

const functionNamespace string = "default"

// MakeDeploy creates a handler for deploying functions
func MakeDeploy(client nomad.Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		req := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create job /v1/jobs
		_, _, err = client.Register(createJob(req), nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			log.Println(err)
			return
		}
	}
}

func createJob(r requests.CreateFunctionRequest) *api.Job {
	jobname := nomad.JobPrefix + r.Service
	job := api.NewServiceJob(jobname, jobname, "global", 1)
	job.Datacenters = []string{"dc1"}

	task := &api.Task{
		Name:   r.Service,
		Driver: "docker",
		Config: map[string]interface{}{
			"image": r.Image,
			"port_map": []map[string]interface{}{
				map[string]interface{}{"http": 8080},
			},
		},
		Resources: &api.Resources{
			Networks: []*api.NetworkResource{
				&api.NetworkResource{
					DynamicPorts: []api.Port{api.Port{Label: "http"}},
				},
			},
		},
		Services: []*api.Service{
			&api.Service{
				Name:      r.Service,
				PortLabel: "http",
			},
		},
	}

	count := 1

	tg := []*api.TaskGroup{
		&api.TaskGroup{
			Name:  &r.Service,
			Count: &count,
			Tasks: []*api.Task{task},
		},
	}

	job.TaskGroups = tg

	return job
}
