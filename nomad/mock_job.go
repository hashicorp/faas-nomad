package nomad

import (
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/mock"
)

// MockJob is a mock implementation of the Job interface
type MockJob struct {
	mock.Mock
}

// Register is a mock implementaion of the register interface method
func (m *MockJob) Register(job *api.Job, options *api.WriteOptions) (
	*api.JobRegisterResponse,
	*api.WriteMeta, error) {

	args := m.Called(job, options)

	var resp *api.JobRegisterResponse
	if r := args.Get(0); r != nil {
		resp = r.(*api.JobRegisterResponse)
	}

	var meta *api.WriteMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.WriteMeta)
	}

	return resp, meta, args.Error(2)

}

// Info returns mock info from the job API
func (m *MockJob) Info(jobID string, q *api.QueryOptions) (*api.Job, *api.QueryMeta, error) {
	args := m.Called(jobID, q)

	var job *api.Job
	if j := args.Get(0); j != nil {
		job = j.(*api.Job)
	}

	var meta *api.QueryMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.QueryMeta)
	}

	return job, meta, args.Error(2)
}

// List returns mock info from the job API
func (m *MockJob) List(q *api.QueryOptions) ([]*api.JobListStub, *api.QueryMeta, error) {
	args := m.Called(q)

	var jobs []*api.JobListStub
	if j := args.Get(0); j != nil {
		jobs = j.([]*api.JobListStub)
	}

	var meta *api.QueryMeta
	if r := args.Get(1); r != nil {
		meta = r.(*api.QueryMeta)
	}

	return jobs, meta, args.Error(2)
}

// Deregister is a mock implementation of the interface method
func (m *MockJob) Deregister(jobID string, purge bool, q *api.WriteOptions) (
	string, *api.WriteMeta, error) {

	args := m.Called(jobID, purge, q)

	return "", nil, args.Error(2)
}
