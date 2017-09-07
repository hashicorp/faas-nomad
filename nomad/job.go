package nomad

import "github.com/hashicorp/nomad/api"

// Job defines the interface for creating a new new job
type Job interface {
	// Register creates a new Nomad job
	Register(*api.Job, *api.WriteOptions) (*api.JobRegisterResponse, *api.WriteMeta, error)
	Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error)
}
