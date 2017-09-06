package nomad

import "github.com/hashicorp/nomad/api"

// Deployments is an interface for the Nomad API
type Deployments interface {
	// List is used to dump all of the evaluations.
	List(q *api.QueryOptions) ([]*api.Deployment, *api.QueryMeta, error)
}
