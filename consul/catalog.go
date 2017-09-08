package consul

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul-template/dependency"
	"github.com/hashicorp/consul-template/watch"
	"github.com/hashicorp/consul/api"
)

// Catalog defines methods for Consul's service catalog
type Catalog interface {
	Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

type ServiceResolver interface {
	Resolve(function string) (string, error)
}

type ConsulResolver struct {
	clientSet  *dependency.ClientSet
	watcher    *watch.Watcher
	deps       []dependency.Dependency
	cache      map[string][]string
	cacheMutex sync.Mutex
}

func NewConsulResolver(address string) *ConsulResolver {
	clientSet := dependency.NewClientSet()
	clientSet.CreateConsulClient(&dependency.CreateConsulClientInput{
		Address: address,
	})

	watch, _ := watch.NewWatcher(&watch.NewWatcherInput{
		Clients:  clientSet,
		MaxStale: 10000 * time.Millisecond,
	})

	cr := &ConsulResolver{
		clientSet:  clientSet,
		watcher:    watch,
		cache:      make(map[string][]string),
		cacheMutex: sync.Mutex{},
	}

	cr.watch()

	return cr
}

func (sr *ConsulResolver) watch() {
	go func() {
		dc := sr.watcher.DataCh()
		for w := range dc {
			log.Println("Service catalog updated", w.Data())

			cs := w.Data().([]*dependency.CatalogService)

			sr.cacheMutex.Lock()
			defer sr.cacheMutex.Unlock()

			sr.cache[cs[0].ServiceName] = make([]string, 0)

			for _, addr := range cs {
				sr.cache[addr.ServiceName] = append(
					sr.cache[addr.ServiceName],
					fmt.Sprintf("http://%v:%v", addr.Address, addr.ServicePort),
				)
			}
		}
	}()
}

func (sr *ConsulResolver) Resolve(function string) ([]string, error) {
	//check the cache
	if val, ok := sr.cache[function]; ok {
		log.Println("Got Address from cache")
		return val, nil
	}

	log.Println("Getting Address from consul")
	q, err := dependency.NewCatalogServiceQuery(function)
	if err != nil {
		return nil, err
	}

	sr.deps = append(sr.deps, q)
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
	sr.cacheMutex.Lock()
	defer sr.cacheMutex.Unlock()

	sr.cache[function] = addresses

	return addresses, nil
}
