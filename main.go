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

var version = "notset"

func setupLogging() hclog.Logger {
	logJSON := false
	if os.Getenv("logger_format") == "json" {
		logJSON = true
	}

	logLevel := "INFO"
	if level := os.Getenv("logger_level"); level != "" {
		logLevel = level
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       "nomadd",
		Level:      hclog.LevelFromString(logLevel),
		JSONFormat: logJSON,
		Output:     createLogFile(),
	})

	return appLogger
}

func createLogFile() *os.File {
	if logFile := os.Getenv("logger_output"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			return f
		}

		log.Printf("Unable to open file for output, defaulting to std out: %s\n", err.Error())
	}

	return os.Stdout
}

func main() {

	logger, stats, nomadClient, consulResolver := makeDependencies(
		os.Getenv("STATSD_ADDR"),
		os.Getenv("NOMAD_ADDR_http"),
		os.Getenv("NOMAD_ADDR"),
		os.Getenv("CONSUL_ADDR"),
		os.Getenv("NOMAD_REGION"),
	)

	logger.Info("Started version: " + version)
	stats.Incr("started", nil, 1)

	handlers := &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(nomadClient.Jobs(), logger, stats),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs(), logger, stats),
		DeleteHandler:  handlers.MakeDelete(consulResolver, nomadClient.Jobs(), logger, stats),
		ReplicaReader:  makeReplicationReader(nomadClient.Jobs(), logger, stats),
		ReplicaUpdater: makeReplicationUpdater(nomadClient.Jobs(), logger, stats),
		FunctionProxy:  makeFunctionProxyHandler(consulResolver, logger, stats),
	}

	config := &types.FaaSConfig{}

	bootstrap.Serve(handlers, config)
}

func makeDependencies(statsDAddr, thisAddr, nomadAddr, consulAddr, region string) (hclog.Logger, *statsd.Client, *api.Client, *consul.Resolver) {
	logger := setupLogging()

	logger.Info("Using StatsD server:" + statsDAddr)
	stats, err := statsd.New(statsDAddr)
	if err != nil {
		logger.Error("Error creating statsd client", err)
	}

	// prefix every metric with the app name
	stats.Namespace = "faas.nomadd."
	stats.Tags = append(stats.Tags, "instance:"+strings.Replace(thisAddr, ":", "_", -1))

	c := api.DefaultConfig()
	nomadClient, err := api.NewClient(c.ClientConfig(region, nomadAddr, false))
	if err != nil {
		logger.Error("Unable to create nomad client", err)
	}

	cr := consul.NewResolver(consulAddr)

	return logger, stats, nomadClient, cr
}

func makeFunctionProxyHandler(r consul.ServiceResolver, logger hclog.Logger, s *statsd.Client) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeProxy(handlers.MakeProxyClient(), r, logger, s),
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
