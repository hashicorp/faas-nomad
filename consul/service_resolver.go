package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul-template/dependency"
	"github.com/hashicorp/consul-template/watch"
	"github.com/hashicorp/consul/api"
	hclog "github.com/hashicorp/go-hclog"

	cache "github.com/patrickmn/go-cache"
)

// Catalog defines methods for Consul's service catalog
type Catalog interface {
	Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

// CatalogServiceQuery defines an interface for Consul Template service query
type CatalogServiceQuery interface {
	Fetch(clients *dependency.ClientSet, opts *dependency.QueryOptions) (interface{}, *dependency.ResponseMetadata, error)
	CanShare() bool
	Stop()
	String() string
	Type() dependency.Type
}

// WrappedWatcher wraps watch.View to allow testing
type WrappedWatcher struct {
	*watch.Watcher
}

type iterateFunc func(dep dependency.Dependency, deps []*dependency.CatalogService)

// IterateDataCh returns the list of CatalogService from the View
func (ww *WrappedWatcher) IterateDataCh(f iterateFunc) {
	for cs := range ww.DataCh() {
		f(
			cs.Dependency(),
			cs.Data().([]*dependency.CatalogService),
		)
	}
}

// Watcher is an interface to the Consul Template watcher struct
type Watcher interface {
	Add(dependency dependency.Dependency) (bool, error)
	Remove(dependency dependency.Dependency) bool
	IterateDataCh(iterateFunc)
}

// ServiceResolver uses consul to resolve a function name into addresses
type ServiceResolver interface {
	Resolve(function string) ([]string, error)
	RemoveCacheItem(service string)
}

// Resolver implements ServiceResolver
type Resolver struct {
	clientSet       *dependency.ClientSet
	watcher         Watcher
	cache           *cache.Cache
	getServiceQuery func(service string) (CatalogServiceQuery, error)
	logger          hclog.Logger
}

type cacheItem struct {
	serviceQuery dependency.Dependency
	addresses    []string
}

// NewResolver creates a new Resolver
func NewResolver(address, ACLToken string, logger hclog.Logger) *Resolver {
	clientSet := dependency.NewClientSet()
	clientSet.CreateConsulClient(&dependency.CreateConsulClientInput{
		Address: address,
		Token:   ACLToken,
	})

	watch, _ := watch.NewWatcher(&watch.NewWatcherInput{
		Clients:  clientSet,
		MaxStale: 10000 * time.Millisecond,
	})

	pc := cache.New(5*time.Minute, 10*time.Minute)

	cr := &Resolver{
		clientSet:       clientSet,
		watcher:         &WrappedWatcher{watch},
		cache:           pc,
		getServiceQuery: createServiceQueryImpl,
		logger:          logger,
	}

	go cr.watch()

	return cr
}

// createServiceQueryImpl allows the mocking of the process to create a consul service query
func createServiceQueryImpl(function string) (CatalogServiceQuery, error) {
	return dependency.NewCatalogServiceQuery(function)
}

// Resolve resolves a function name to an array of URI
func (sr *Resolver) Resolve(function string) ([]string, error) {
	//check the cache
	if val, ok := sr.cache.Get(getCacheKey(function)); ok {
		sr.logger.Info("Got Address from cache", "address", fmt.Sprintf("%#v", val))
		return val.(*cacheItem).addresses, nil
	}

	sr.logger.Info("Getting Address from consul", "function", function)
	q, err := sr.getServiceQuery(function)
	if err != nil {
		return nil, err
	}

	s, _, err := q.Fetch(sr.clientSet, nil)
	if err != nil {
		return nil, err
	}
	sr.watcher.Add(q)

	cs := s.([]*dependency.CatalogService)
	addresses := sr.updateCatalog(q, cs)

	return addresses, nil
}

// RemoveCacheItem removes a service reference from the cache
func (sr *Resolver) RemoveCacheItem(function string) {
	key := getCacheKey(function)
	if d, ok := sr.cache.Get(key); ok {
		sr.watcher.Remove(d.(*cacheItem).serviceQuery)
		sr.cache.Delete(key)
	}
}

// watch watches consul for changes and updates the cache on change
func (sr *Resolver) watch() {
	sr.watcher.IterateDataCh(
		func(dep dependency.Dependency, deps []*dependency.CatalogService) {
			sr.updateCatalog(dep, deps)
		},
	)
}

func (sr *Resolver) updateCatalog(dep dependency.Dependency, cs []*dependency.CatalogService) []string {
	sr.logger.Info("Service catalog updated", "dependency", dep.String())
	addresses := make([]string, 0)

	if len(cs) < 1 {
		sr.upsertCache(dep, addresses)
		return addresses
	}

	for _, addr := range cs {
		addresses = append(
			addresses,
			fmt.Sprintf("http://%v:%v", addr.ServiceAddress, addr.ServicePort),
		)
	}

	sr.upsertCache(dep, addresses)

	return addresses
}

func (sr *Resolver) upsertCache(dep dependency.Dependency, value []string) {
	if ci, ok := sr.cache.Get(dep.String()); ok {
		ci.(*cacheItem).addresses = value
		sr.cache.Set(dep.String(), ci, cache.NoExpiration)

		return
	}

	sr.cache.Set(dep.String(), &cacheItem{
		addresses:    value,
		serviceQuery: dep,
	}, cache.NoExpiration)
}

func getCacheKey(function string) string {
	return fmt.Sprintf("catalog.service(%s)", function)
}
