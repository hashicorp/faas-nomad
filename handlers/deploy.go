package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

var (
	count           = 1
	restartDelay    = 1 * time.Second
	restartMode     = "delay"
	restartAttempts = 25
	logFiles        = 5
	logSize         = 2

	// Constraints
	constraintCPUArch = "amd64"
	taskMemory        = 128

	// Update Strategy
	updateAutoRevert      = true
	updateMinHealthyTime  = 5 * time.Second
	updateHealthyDeadline = 20 * time.Second
	updateStagger         = 5 * time.Second
)

// MakeDeploy creates a handler for deploying functions
func MakeDeploy(client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	log := logger.Named("deploy_handler")

	return func(w http.ResponseWriter, r *http.Request) {
		stats.Incr("deploy.called", nil, 1)

		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		req := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			stats.Incr("deploy.error.badrequest", nil, 1)
			return
		}

		// Create job /v1/jobs
		_, _, err = client.Register(createJob(req), nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))

			log.Error("Error registering job", "error", err.Error())
			stats.Incr("deploy.error.createjob", []string{"job:" + req.Service}, 1)
			return
		}

		stats.Incr("deploy.success", []string{"job:" + req.Service}, 1)
		stats.Gauge("deploy.count", 1, []string{"job:" + req.Service}, 1)
	}
}

func createJob(r requests.CreateFunctionRequest) *api.Job {
	jobname := nomad.JobPrefix + r.Service
	job := api.NewServiceJob(jobname, jobname, "global", 1)
	count := 1
	restartDelay := 1 * time.Second
	restartMode := "delay"
	restartAttempts := 25
	taskMemory := 128
	taskCPU := 100
	logFiles := 5
	logSize := 2
	envVars := r.EnvVars

	if r.Labels != nil && (*r.Labels)["datacenters"] != "" {
		lbls := (*r.Labels)["datacenters"]
		dcs := []string{}

		for _, dc := range strings.Split(lbls, ",") {
			dcs = append(dcs, strings.Trim(dc, " "))
		}

		job.Datacenters = dcs
	} else {
		// default datacenter
		job.Datacenters = []string{"dc1"}
	}

	if r.Limits != nil {
		if r.Limits.CPU != "" {
			v, err := strconv.ParseInt(r.Limits.CPU, 10, 32)
			if err == nil {
				taskCPU = int(v)
			}
		}

		if r.Limits.Memory != "" {
			v, err := strconv.ParseInt(r.Limits.Memory, 10, 32)
			if err == nil {
				taskMemory = int(v)
			}
		}
	}

	if envVars == nil {
		envVars = map[string]string{}
	}

	// add constraints
	job.Constraints = append(job.Constraints,
		&api.Constraint{
			LTarget: "${attr.cpu.arch}",
			Operand: "=",
			RTarget: constraintCPUArch,
		},
	)

	// add rolling update
	job.Update = &api.UpdateStrategy{
		MinHealthyTime:  &updateMinHealthyTime,
		AutoRevert:      &updateAutoRevert,
		Stagger:         &updateStagger,
		HealthyDeadline: &updateHealthyDeadline,
	}

	if r.EnvProcess != "" {
		envVars["fprocess"] = r.EnvProcess
	}

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
			MemoryMB: &taskMemory,
			CPU:      &taskCPU,
		},
		Services: []*api.Service{
			&api.Service{
				Name:      r.Service,
				PortLabel: "http",
			},
		},
		LogConfig: &api.LogConfig{
			MaxFiles:      &logFiles,
			MaxFileSizeMB: &logSize,
		},
		Env: envVars,
	}

	tg := []*api.TaskGroup{
		&api.TaskGroup{
			Name:  &r.Service,
			Count: &count,
			RestartPolicy: &api.RestartPolicy{
				Delay:    &restartDelay,
				Mode:     &restartMode,
				Attempts: &restartAttempts,
			},
			Tasks: []*api.Task{task},
		},
	}

	job.TaskGroups = tg

	return job
}
