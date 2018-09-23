package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas/gateway/requests"
)

var (
	count             = 1
	restartDelay      = 1 * time.Second
	restartMode       = "delay"
	restartAttempts   = 25
	logFiles          = 5
	logSize           = 2
	ephemeralDiskSize = 20

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

	job.Meta = createAnnotations(r)
	// job.Constraints = createConstraints(r)
	job.Update = createUpdateStrategy()

	// add constraints
	job.Constraints = append(job.Constraints,
		&api.Constraint{
			LTarget: "${attr.cpu.arch}",
			Operand: "=",
			RTarget: constraintCPUArch,
		},
	)

	job.TaskGroups = createTaskGroup(r)

	return job
}

func createTaskGroup(r requests.CreateFunctionRequest) []*api.TaskGroup {
	count := 1
	restartDelay := 1 * time.Second
	restartMode := "delay"
	restartAttempts := 25
	task := createTask(r)

	return []*api.TaskGroup{
		&api.TaskGroup{
			Name:  &r.Service,
			Count: &count,
			RestartPolicy: &api.RestartPolicy{
				Delay:    &restartDelay,
				Mode:     &restartMode,
				Attempts: &restartAttempts,
			},
			EphemeralDisk: &api.EphemeralDisk{
				SizeMB: &ephemeralDiskSize,
			},
			Tasks: []*api.Task{task},
		},
	}
}

func createTask(r requests.CreateFunctionRequest) *api.Task {
	envVars := createEnvVars(r)

	return &api.Task{
		Name:   r.Service,
		Driver: "docker",
		Config: map[string]interface{}{
			"image": r.Image,
			"port_map": []map[string]interface{}{
				map[string]interface{}{"http": 8080},
			},
			"labels": createLabels(r),
		},
		Resources: createResources(r),
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
}

func createAnnotations(r requests.CreateFunctionRequest) map[string]string {
	annotations := map[string]string{}
	if r.Annotations != nil {
		for k, v := range *r.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}

func createLabels(r requests.CreateFunctionRequest) []map[string]interface{} {
	labels := []map[string]interface{}{}
	if r.Labels != nil {
		for k, v := range *r.Labels {
			labels = append(labels, map[string]interface{}{k: v})
		}
	}
	return labels
}

func createResources(r requests.CreateFunctionRequest) *api.Resources {
	taskMemory, taskCPU := createLimits(r)

	return &api.Resources{
		Networks: []*api.NetworkResource{
			&api.NetworkResource{
				DynamicPorts: []api.Port{api.Port{Label: "http"}},
			},
		},
		MemoryMB: &taskMemory,
		CPU:      &taskCPU,
	}
}

func createLimits(r requests.CreateFunctionRequest) (taskMemory, taskCPU int) {
	taskMemory = 128
	taskCPU = 100

	if r.Limits == nil {
		return taskMemory, taskCPU
	}

	cpu, err := strconv.ParseInt(r.Limits.CPU, 10, 32)
	if err == nil {
		taskCPU = int(cpu)
	}

	mem, err := strconv.ParseInt(r.Limits.Memory, 10, 32)
	if err == nil {
		taskMemory = int(mem)
	}

	return taskMemory, taskCPU
}

func createDataCenters(datacenters string) []string {
	if datacenters != "" {
		dcs := []string{}
		var parseDcs = strings.Split(datacenters, " = ")
		for _, dc := range strings.Split(parseDcs[1], ",") {
			dcs = append(dcs, dc)
		}
		return dcs
	}

	// default datacenter
	return []string{"dc1"}
}

func createConstraints(r requests.CreateFunctionRequest) []string {
	if r.Constraints != nil {
		for _, constraint := range r.Constraints {
			var parseConstraint = strings.Split(constraint, " = ")
			if parseConstraint[0] == "datacenters" {
				createDataCenters(constraint)
			}
		}
	}
	return []string{}
}

func createEnvVars(r requests.CreateFunctionRequest) map[string]string {
	envVars := map[string]string{}

	if r.EnvVars != nil {
		envVars = r.EnvVars
	}

	if r.EnvProcess != "" {
		envVars["fprocess"] = r.EnvProcess
	}

	return envVars
}

func createUpdateStrategy() *api.UpdateStrategy {
	return &api.UpdateStrategy{
		MinHealthyTime:  &updateMinHealthyTime,
		AutoRevert:      &updateAutoRevert,
		Stagger:         &updateStagger,
		HealthyDeadline: &updateHealthyDeadline,
	}
}
