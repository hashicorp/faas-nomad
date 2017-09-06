package nomad

import (
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/mock"
)

type MockDeployments struct {
	mock.Mock
}

func (m *MockDeployments) List(q *api.QueryOptions) ([]*api.Deployment, *api.QueryMeta, error) {
	args := m.Called(q)

	var md *api.QueryMeta
	if m := args.Get(1); m != nil {
		md = m.(*api.QueryMeta)
	}

	return args.Get(0).([]*api.Deployment), md, args.Error(2)
}
