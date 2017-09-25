package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/handlers"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/types"
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
		FunctionReader: handlers.MakeReader(nomadClient.Jobs()),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs()),
		DeleteHandler:  handlers.MakeDelete(nomadClient.Jobs()),
		ReplicaReader:  makeReplicationReader(nomadClient.Jobs()),
		ReplicaUpdater: makeReplicationUpdater(nomadClient.Jobs()),
		FunctionProxy:  makeFunctionProxyHandler(r),
	}
	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}

func makeFunctionProxyHandler(r consul.ServiceResolver) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeProxy(handlers.MakeProxyClient(), r),
	)
}

func makeReplicationReader(client nomad.Job) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeReplicationReader(client),
	)
}

func makeReplicationUpdater(client nomad.Job) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeReplicationWriter(client),
	)
}
