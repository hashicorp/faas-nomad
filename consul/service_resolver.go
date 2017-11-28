package consul

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul-template/dependency"
	"github.com/hashicorp/consul-template/watch"
	"github.com/hashicorp/consul/api"
	cache "github.com/patrickmn/go-cache"
)

// Catalog defines methods for Consul's service catalog
type Catalog interface {
	Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

// ServiceResolver uses consul to resolve a function name into addresses
type ServiceResolver interface {
	Resolve(function string) ([]string, error)
	RemoveCacheItem(service string)
}

// Resolver implements ServiceResolver
type Resolver struct {
	clientSet *dependency.ClientSet
	watcher   *watch.Watcher
	cache     *cache.Cache
}

type cacheItem struct {
	serviceQuery dependency.Dependency
	addresses    []string
}

// NewResolver creates a new Resolver
func NewResolver(address string) *Resolver {
	clientSet := dependency.NewClientSet()
	clientSet.CreateConsulClient(&dependency.CreateConsulClientInput{
		Address: address,
	})

	watch, _ := watch.NewWatcher(&watch.NewWatcherInput{
		Clients:  clientSet,
		MaxStale: 10000 * time.Millisecond,
	})

	pc := cache.New(5*time.Minute, 10*time.Minute)

	cr := &Resolver{
		clientSet: clientSet,
		watcher:   watch,
		cache:     pc,
	}

	cr.watch()

	return cr
}

// Resolve resolves a function name to an array of URI
func (sr *Resolver) Resolve(function string) ([]string, error) {
	//check the cache
	if val, ok := sr.cache.Get(function); ok {
		log.Println("Got Address from cache")
		return val.(*cacheItem).addresses, nil
	}

	log.Println("Getting Address from consul")
	q, err := dependency.NewCatalogServiceQuery(function)
	if err != nil {
		return nil, err
	}

	sr.watcher.Add(q)

	s, _, err := q.Fetch(sr.clientSet, nil)
	if err != nil {
		return nil, err
	}

	cs := s.([]*dependency.CatalogService)
	addresses := make([]string, 0)

	for _, a := range cs {
		addresses = append(addresses, fmt.Sprintf("http://%v:%v", a.Address, a.ServicePort))
	}

	// append the cache
	ci := &cacheItem{
		addresses:    addresses,
		serviceQuery: q,
	}
	sr.cache.Set(function, ci, cache.DefaultExpiration)

	return addresses, nil
}

// RemoveCacheItem removes a service reference from the cache
func (sr *Resolver) RemoveCacheItem(function string) {
	if d, ok := sr.cache.Get(function); ok {
		sr.watcher.Remove(d.(*cacheItem).serviceQuery)
		sr.cache.Delete(function)
	}
}

// watch watches consul for changes and updates the cache on change
func (sr *Resolver) watch() {
	go func() {
		dc := sr.watcher.DataCh()
		for w := range dc {
			log.Println("Service catalog updated", w.Data())

			cs := w.Data().([]*dependency.CatalogService)

			addresses := make([]string, 0)
			for _, addr := range cs {
				addresses = append(
					addresses,
					fmt.Sprintf("http://%v:%v", addr.Address, addr.ServicePort),
				)
			}

			if len(cs) > 0 {
				ci, ok := sr.cache.Get(cs[0].ServiceName)
				if ok {
					ci.(*cacheItem).addresses = addresses
					sr.cache.Set(cs[0].ServiceName, ci, cache.DefaultExpiration)
				}
			}
		}
	}()
}
