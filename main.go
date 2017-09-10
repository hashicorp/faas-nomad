package main

import (
	"log"
	"os"

	bootstrap "github.com/alexellis/faas-provider"
	"github.com/alexellis/faas-provider/types"
	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/handlers"
	"github.com/hashicorp/nomad/api"
)

func main() {
	region := os.Getenv("NOMAD_REGION")
	nomadAddr := os.Getenv("NOMAD_ADDR")
	consulAddr := os.Getenv("CONSUL_ADDR")

	c := api.DefaultConfig()

	nomadClient, err := api.NewClient(c.ClientConfig(region, nomadAddr, false))
	if err != nil {
		log.Fatal(err)
	}

	r := consul.NewConsulResolver(consulAddr)

	handlers := &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(nomadClient.Allocations()),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs()),
		DeleteHandler:  handlers.MakeNull(),
		ReplicaReader:  handlers.MakeNull(),
		FunctionProxy:  handlers.MakeProxy(handlers.MakeProxyClient(), r),
		ReplicaUpdater: handlers.MakeNull(),
	}
	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}
