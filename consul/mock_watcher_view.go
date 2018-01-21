package consul

import (
	"github.com/hashicorp/consul-template/dependency"
	"github.com/hashicorp/consul-template/watch"
)

type MockWatcherView struct {
	*watch.View
	data []*dependency.CatalogService
}

func (w *MockWatcherView) Data() interface{} {
	return w.data
}
