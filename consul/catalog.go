package consul

import "github.com/hashicorp/consul/api"

// Catalog defines methods for Consul's service catalog
type Catalog interface {
	Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}
