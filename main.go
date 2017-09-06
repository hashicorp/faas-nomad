package main

import (
	bootstrap "github.com/alexellis/faas-provider"
	"github.com/alexellis/faas-provider/types"
	"github.com/nicholasjackson/faas-nomad/handlers"
)

func main() {
	handlers := &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(),
		DeployHandler:  handlers.MakeReader(),
		DeleteHandler:  handlers.MakeReader(),
		ReplicaReader:  handlers.MakeReader(),
		FunctionProxy:  handlers.MakeReader(),
		ReplicaUpdater: handlers.MakeReader(),
	}
	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}
