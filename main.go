package main

import (
	"log"

	bootstrap "github.com/alexellis/faas-provider"
	"github.com/alexellis/faas-provider/types"
	"github.com/hashicorp/nomad/api"
	"github.com/nicholasjackson/faas-nomad/handlers"
)

func main() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	handlers := &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(client.Deployments()),
		DeployHandler:  handlers.MakeReader(client.Deployments()),
		DeleteHandler:  handlers.MakeReader(client.Deployments()),
		ReplicaReader:  handlers.MakeReader(client.Deployments()),
		FunctionProxy:  handlers.MakeReader(client.Deployments()),
		ReplicaUpdater: handlers.MakeReader(client.Deployments()),
	}
	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}
