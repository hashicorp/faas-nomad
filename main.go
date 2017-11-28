package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/gorilla/mux"
	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/handlers"
	"github.com/hashicorp/faas-nomad/metrics"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/types"
)

const version = "0.2.8"

func main() {
	region := os.Getenv("NOMAD_REGION")
	nomadAddr := os.Getenv("NOMAD_ADDR")
	consulAddr := os.Getenv("CONSUL_ADDR")
	statsDAddr := os.Getenv("STATSD_ADDR")
	thisAddr := os.Getenv("NOMAD_ADDR_http")

	log.Println("Started version:", version)
	log.Println("Using StatsD server:", statsDAddr)

	stats, err := statsd.New(statsDAddr)
	if err != nil {
		log.Println(err)
	}
	// prefix every metric with the app name
	stats.Namespace = "faas.nomadd."
	stats.Tags = append(stats.Tags, "instance:"+strings.Replace(thisAddr, ":", "_", -1))

	err = stats.Incr("started", nil, 1)
	if err != nil {
		log.Println(err)
	}

	c := api.DefaultConfig()

	nomadClient, err := api.NewClient(c.ClientConfig(region, nomadAddr, false))
	if err != nil {
		log.Fatal(err)
	}

	cr := consul.NewResolver(consulAddr)

	handlers := &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(nomadClient.Jobs(), stats),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs(), stats),
		DeleteHandler:  handlers.MakeDelete(cr, nomadClient.Jobs(), stats),
		ReplicaReader:  makeReplicationReader(nomadClient.Jobs(), stats),
		ReplicaUpdater: makeReplicationUpdater(nomadClient.Jobs(), stats),
		FunctionProxy:  makeFunctionProxyHandler(cr, statsDAddr, stats),
	}

	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}

func makeFunctionProxyHandler(r consul.ServiceResolver, statsDAddr string, s *statsd.Client) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeProxy(handlers.MakeProxyClient(), r, statsDAddr, s),
	)
}

func makeReplicationReader(client nomad.Job, stats metrics.StatsD) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeReplicationReader(client, stats),
	)
}

func makeReplicationUpdater(client nomad.Job, stats metrics.StatsD) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeReplicationWriter(client, stats),
	)
}
