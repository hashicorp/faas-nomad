package nomad

import "github.com/hashicorp/nomad/api"

// JobPrefix contains a string that prefixes all OpenFaaS jobs
const JobPrefix = "OpenFaaS-"

// Job defines the interface for creating a new new job
type Job interface {
	// Register creates a new Nomad job
	Register(*api.Job, *api.WriteOptions) (*api.JobRegisterResponse, *api.WriteMeta, error)
	Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error)
	List(q *api.QueryOptions) ([]*api.JobListStub, *api.QueryMeta, error)
	Deregister(jobID string, purge bool, q *api.WriteOptions) (string, *api.WriteMeta, error)
}
