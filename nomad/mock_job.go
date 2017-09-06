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
