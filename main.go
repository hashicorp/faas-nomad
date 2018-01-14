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
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/types"
)

const version = "test build"

func setupLogging() hclog.Logger {
	logJSON := false
	if os.Getenv("logger_format") == "json" {
		logJSON = true
	}

	logLevel := "INFO"
	if level := os.Getenv("logger_level"); level != "" {
		logLevel = level
	}

	output := os.Stdout
	if logFile := os.Getenv("logger_output"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			output = f
		} else {
			log.Printf("Unable to open file for output, defaulting to std out: %s\n", err.Error())
		}
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       "nomadd",
		Level:      hclog.LevelFromString(logLevel),
		JSONFormat: logJSON,
		Output:     output,
	})

	return appLogger
}

func main() {
	region := os.Getenv("NOMAD_REGION")
	nomadAddr := os.Getenv("NOMAD_ADDR")
	consulAddr := os.Getenv("CONSUL_ADDR")
	statsDAddr := os.Getenv("STATSD_ADDR")
	thisAddr := os.Getenv("NOMAD_ADDR_http")

	logger := setupLogging()

	logger.Info("Started version: " + version)
	logger.Info("Using StatsD server:" + statsDAddr)

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
		FunctionReader: handlers.MakeReader(nomadClient.Jobs(), logger, stats),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs(), logger, stats),
		DeleteHandler:  handlers.MakeDelete(cr, nomadClient.Jobs(), logger, stats),
		ReplicaReader:  makeReplicationReader(nomadClient.Jobs(), logger, stats),
		ReplicaUpdater: makeReplicationUpdater(nomadClient.Jobs(), logger, stats),
		FunctionProxy:  makeFunctionProxyHandler(cr, statsDAddr, logger, stats),
	}

	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}

func makeFunctionProxyHandler(r consul.ServiceResolver, statsDAddr string, logger hclog.Logger, s *statsd.Client) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeProxy(handlers.MakeProxyClient(), r, statsDAddr, logger, s),
	)
}

func makeReplicationReader(client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeReplicationReader(client, logger, stats),
	)
}

func makeReplicationUpdater(client nomad.Job, logger hclog.Logger, stats metrics.StatsD) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeReplicationWriter(client, logger, stats),
	)
}
