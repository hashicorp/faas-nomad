package nomad

import "github.com/hashicorp/nomad/api"

// Allocations is an interface for the Nomad API
type Allocations interface {
	// List is used to dump all of the evaluations.
	List(q *api.QueryOptions) ([]*api.AllocationListStub, *api.QueryMeta, error)
	Info(allocID string, q *api.QueryOptions) (*api.Allocation, *api.QueryMeta, error)
}
