package handlers

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	fntypes "github.com/hashicorp/faas-nomad/types"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDeploy(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	mockJob.On("Register", mock.Anything, mock.Anything).Return(nil, nil, nil)

	mockStats := &metrics.MockStatsD{}
	mockStats.On("Incr", mock.Anything, mock.Anything, mock.Anything)
	mockStats.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	logger := hclog.Default()

	return MakeDeploy(mockJob, fntypes.ProviderConfig{Vault: fntypes.VaultConfig{DefaultPolicy: "openfaas", SecretPathPrefix: "secret/openfaas"}, Datacenter: "dc1", ConsulAddress: "http://localhost:8500", ConsulDNSEnabled: true}, logger, mockStats),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/functions", bytes.NewReader([]byte(body)))
}

func TestHandlerReturnsErrorOnInvalidRequest(t *testing.T) {
	h, rw, r := setupDeploy("")

	h(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestHandlerRegistersJob(t *testing.T) {
	h, rw, r := setupDeploy(createRequest().String())

	h(rw, r)

	mockJob.AssertCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestHandlerRegistersWithEnvironmentVariables(t *testing.T) {
	reqEnv := map[string]string{"VAR1": "ABC"}
	fr := createRequest()
	fr.EnvVars = reqEnv

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	jobEnv := job.TaskGroups[0].Tasks[0].Env

	assert.Equal(t, reqEnv, jobEnv)
}

func TestHandlerRegistersWithFunctionProcess(t *testing.T) {
	fr := createRequest()
	fr.EnvProcess = "env"

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	jobEnv := job.TaskGroups[0].Tasks[0].Env

	assert.Equal(t, "env", jobEnv["fprocess"])
}

func TestHandlesDataCentreLabelWithSingleDC(t *testing.T) {
	fr := createRequest()
	fr.Constraints = []string{"datacenter == test"}

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	dcs := job.Datacenters

	assert.Equal(t, "test", dcs[0])
}

func TestHandlesDataCentreLabelWithMultipleDC(t *testing.T) {
	fr := createRequest()
	fr.Constraints = []string{"datacenter == test1", "datacenter == test2"}

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	dcs := job.Datacenters

	assert.Equal(t, "test1", dcs[0])
	assert.Equal(t, "test2", dcs[1])
}

func TestHandlesBlankDataCentreLabel(t *testing.T) {
	fr := createRequest()

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	dcs := job.Datacenters

	assert.Equal(t, "dc1", dcs[0])
}

func TestHandlesRequestWithCPULimit(t *testing.T) {
	fr := createRequest()
	fr.Limits = &requests.FunctionResources{
		CPU: "1000",
	}

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	cpu := job.TaskGroups[0].Tasks[0].Resources.CPU

	assert.Equal(t, 1000, *cpu)
}

func TestHandlesRequestWithMemoryLimit(t *testing.T) {
	fr := createRequest()
	fr.Limits = &requests.FunctionResources{
		Memory: "256",
	}

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	mem := job.TaskGroups[0].Tasks[0].Resources.MemoryMB

	assert.Equal(t, 256, *mem)
}

func TestHandlesRequestWithSecrets(t *testing.T) {
	fr := createRequest()
	fr.Secrets = []string{"figlet"}
	expectedTemplate := `{{with secret "secret/openfaas/figlet"}}{{.Data.value}}{{end}}`

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	templates := job.TaskGroups[0].Tasks[0].Templates

	assert.Equal(t, "secrets/figlet", *templates[0].DestPath)
	assert.Equal(t, expectedTemplate, *templates[0].EmbeddedTmpl)
}

func TestHandlesRequestWithDNSServers(t *testing.T) {
	fr := createRequest()
	expectedServers := []string{"127.0.0.1", "127.0.1.1"}
	fr.EnvVars = map[string]string{"dns_servers": "127.0.0.1,127.0.1.1"}

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)

	dnsServers := job.TaskGroups[0].Tasks[0].Config["dns_servers"].([]string)

	assert.Equal(t, expectedServers, dnsServers)
}

func TestHandlesRequestUsingConsulAddress(t *testing.T) {
	fr := createRequest()
	expectedServers := []string{"localhost"}

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)

	dnsServers := job.TaskGroups[0].Tasks[0].Config["dns_servers"].([]string)

	assert.Equal(t, expectedServers, dnsServers)
}

func TestHandleDeployWithRegistryAuth(t *testing.T) {
	fr := createRequest()
	encoded := base64.StdEncoding.EncodeToString([]byte("username:password"))
	fr.RegistryAuth = encoded

	h, rw, r := setupDeploy(fr.String())

	h(rw, r)

	args := mockJob.Calls[0].Arguments
	job := args.Get(0).(*api.Job)
	auth := job.TaskGroups[0].Tasks[0].Config["auth"].([]map[string]interface{})

	assert.Equal(t, "username", auth[0]["username"])
	assert.Equal(t, "password", auth[0]["password"])
}
