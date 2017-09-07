package main

import (
	"log"
	"os"

	bootstrap "github.com/alexellis/faas-provider"
	"github.com/alexellis/faas-provider/types"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/nomad/api"
	"github.com/nicholasjackson/faas-nomad/handlers"
)

func main() {
	region := os.Getenv("NOMAD_REGION")
	nomadAddr := os.Getenv("NOMAD_ADDR")
	consulAddr := os.Getenv("CONSUL_ADDR")

	c := api.DefaultConfig()

	client, err := api.NewClient(c.ClientConfig(region, nomadAddr, false))
	if err != nil {
		log.Fatal(err)
	}

	cc := consul.DefaultConfig()
	cc.Address = consulAddr
	consulApi, err := consul.NewClient(cc)
	if err != nil {
		log.Fatal(err)
	}

	handlers := &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(client.Allocations()),
		DeployHandler:  handlers.MakeDeploy(client.Jobs()),
		DeleteHandler:  handlers.MakeNull(),
		ReplicaReader:  handlers.MakeNull(),
		FunctionProxy:  handlers.MakeProxy(client.Jobs(), consulApi.Catalog()),
		ReplicaUpdater: handlers.MakeNull(),
	}
	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}
