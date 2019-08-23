package consul

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/consul-template/dependency"
	hclog "github.com/hashicorp/go-hclog"
	cache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setup(t *testing.T, queryError error) (*Resolver, *MockWatcher, *cache.Cache, *MockServiceQuery) {

	dep := &dependency.CatalogService{
		ServiceName:    "test",
		Address:        "myaddress",
		ServiceAddress: "myaddress",
		ServicePort:    8080,
	}
	cs := []*dependency.CatalogService{dep}

	serviceQuery := &MockServiceQuery{name: "test"}
	serviceQuery.On("Fetch", mock.Anything).Return(cs, nil, nil)

	watcher := &MockWatcher{data: make(chan []*dependency.CatalogService)}
	watcher.On("Add", mock.Anything).Return(true, nil)
	watcher.On("Remove", mock.Anything).Return(true)
	watcher.On("IterateDataCh", mock.Anything).Return(serviceQuery)

	pc := cache.New(5*time.Minute, 10*time.Minute)
	return &Resolver{
			cache:   pc,
			watcher: watcher,
			getServiceQuery: func(function string) (CatalogServiceQuery, error) {
				return serviceQuery, queryError
			},
			clientSet: &dependency.ClientSet{},
			logger:    hclog.Default(),
		},
		watcher,
		pc,
		serviceQuery
}

func TestResolveWithCacheReturnsURLS(t *testing.T) {
	address := "http://mytest.com"
	r, _, c, sq := setup(t, nil)
	c.Add(sq.String(), &cacheItem{addresses: []string{address}}, cache.DefaultExpiration)

	addr, err := r.Resolve("test")

	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, addr[0], "http://mytest.com")
}

func TestResolveWithNoCacheQueriesTheCatalog(t *testing.T) {
	r, _, _, _ := setup(t, nil)

	addr, err := r.Resolve("test")

	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, addr[0], "http://myaddress:8080")
}

func TestResolveWithNoCacheAddsAWatcher(t *testing.T) {
	r, w, _, _ := setup(t, nil)

	r.Resolve("test")

	w.AssertCalled(t, "Add", mock.Anything)
}

func TestResolveWithNoCacheUpdatesCache(t *testing.T) {
	r, _, c, q := setup(t, nil)

	r.Resolve("test")

	ci, b := c.Get(q.String())

	assert.True(t, b)
	assert.NotNil(t, ci.(*cacheItem).serviceQuery)
	assert.Equal(t, "http://myaddress:8080", ci.(*cacheItem).addresses[0])
}

func TestResolveWithNoAddressesUpdatesCache(t *testing.T) {
	r, _, c, sq := setup(t, nil)
	sq.ExpectedCalls = make([]*mock.Call, 0)
	sq.On("Fetch", mock.Anything, mock.Anything).Return(make([]*dependency.CatalogService, 0), nil, nil)

	r.Resolve("test")

	ci, b := c.Get(sq.String())

	assert.True(t, b)
	assert.NotNil(t, ci.(*cacheItem).serviceQuery)
	assert.Equal(t, 0, len(ci.(*cacheItem).addresses))
}

func TestResolveWithUnableToCreateQueryReturnsError(t *testing.T) {
	r, _, _, _ := setup(t, fmt.Errorf("Boom"))

	addr, err := r.Resolve("test")

	assert.NotNil(t, err, "Error should not be nil")
	assert.Nil(t, addr, "Address should be nil")
}

func TestResolveWithUnableToFetchReturnsError(t *testing.T) {
	r, w, _, sq := setup(t, nil)
	sq.ExpectedCalls = make([]*mock.Call, 0)
	sq.On("Fetch", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("Boom"))

	addr, err := r.Resolve("test")

	assert.NotNil(t, err, "Error should not be nil")
	assert.Nil(t, addr, "Address should be nil")
	w.AssertNotCalled(t, "Add", mock.Anything, mock.Anything)
}

func TestWatchWithNewServicesUpdatesCatalog(t *testing.T) {
	r, w, _, _ := setup(t, nil)
	go r.watch()

	time.Sleep(1 * time.Millisecond)
	r.Resolve("test")

	w.data <- []*dependency.CatalogService{
		&dependency.CatalogService{
			ServiceName:    "test",
			Address:        "mynewaddress",
			ServiceAddress: "mynewaddress",
			ServicePort:    8081,
		},
	}

	time.Sleep(1 * time.Millisecond)
	addr, err := r.Resolve("test")

	assert.Nil(t, err)
	assert.Equal(t, "http://mynewaddress:8081", addr[0])
}

func TestRemoveCacheItemRemovesFromCache(t *testing.T) {
	r, _, c, _ := setup(t, nil)

	r.Resolve("test")
	r.RemoveCacheItem("test")
	_, ok := c.Get("test")

	assert.False(t, ok)
}
