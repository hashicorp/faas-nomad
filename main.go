package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

var (
	port         = flag.Int("port", 8080, "Port to bind the server to")
	statsdServer = flag.String("statsd_addr", "localhost:8125", "Location for the statsd collector")
	nodeURI      = flag.String("node_addr", "localhost", "URI of the current Nomad node, this address is used for reporting and logging")
	nomadAddr    = flag.String("nomad_addr", "localhost:4646", "Address for Nomad API endpoint")
	consulAddr   = flag.String("consul_addr", "http://localhost:8500", "Address for Consul API endpoint")
	nomadRegion  = flag.String("nomad_region", "global", "Default region to schedule functions in")
)

var functionTimeout = flag.Duration("function_timeout", 30*time.Second, "Timeout for function execution")

var (
	loggerFormat = flag.String("logger_format", "text", "Format for log output text | json")
	loggerLevel  = flag.String("logger_level", "INFO", "Log output level INFO | ERROR | DEBUG | TRACE")
	loggerOutput = flag.String("logger_output", "", "Filepath to write log file, if omitted stdOut is used")
)

// parseDeprecatedEnvironment is used to merge the previous environment variable configuration to the new flag style
// this will be removed in the next release
func parseDeprecatedEnvironment() {
	checkDeprecatedStatsD()
	checkDeprecatedNomadHTTP()
	checkDeprecatedNomadAddr()
	checkDeprecatedConsulAddr()
	checkDeprecatedNomadRegion()
	checkDeprecatedLoggerLevel()
	checkDeprecatedLoggerFormat()
	checkDeprecatedLoggerOutput()
}

func checkDeprecatedStatsD() {
	if env := os.Getenv("STATSD_ADDR"); env != "" {
		*statsdServer = env
		log.Println("The environment variable STATSD_ADDR is deprecated please use the command line flag stasd_server")
	}
}

func checkDeprecatedNomadHTTP() {
	if env := os.Getenv("NOMAD_ADDR_http"); env != "" {
		*nodeURI = env
		log.Println("The environment variable NOMAD_ADDR_http is deprecated please use the command line flag node_uri")
	}
}

func checkDeprecatedNomadAddr() {
	if env := os.Getenv("NOMAD_ADDR"); env != "" {
		*nomadAddr = env
		log.Println("The environment variable NOMAD_ADDR is deprecated please use the command line flag nomad_addr")
	}
}

func checkDeprecatedConsulAddr() {
	if env := os.Getenv("CONSUL_ADDR"); env != "" {
		*consulAddr = env
		log.Println("The environment variable CONSUL_ADDR is deprecated please use the command line flag consul_addr")
	}
}

func checkDeprecatedNomadRegion() {
	if env := os.Getenv("NOMAD_REGION"); env != "" {
		*nomadRegion = env
		log.Println("The environment variable NOMAD_REGION is deprecated please use the command line flag nomad_region")
	}
}

func checkDeprecatedLoggerLevel() {
	if env := os.Getenv("logger_level"); env != "" {
		*loggerLevel = env
		log.Println("The environment variable logger_level is deprecated please use the command line flag logger_level")
	}
}

func checkDeprecatedLoggerFormat() {
	if env := os.Getenv("logger_format"); env != "" {
		*loggerFormat = env
		log.Println("The environment variable logger_format is deprecated please use the command line flag logger_format")
	}
}

func checkDeprecatedLoggerOutput() {
	if env := os.Getenv("logger_output"); env != "" {
		*loggerOutput = env
		log.Println("The environment variable logger_output is deprecated please use the command line flag logger_output")
	}
}

func main() {
	flag.Parse()
	parseDeprecatedEnvironment() // to be removed in 0.3.0

	logger, stats, nomadClient, consulResolver := makeDependencies(
		*statsdServer,
		*nodeURI,
		*nomadAddr,
		*consulAddr,
		*nomadRegion,
	)

	logger.Info("Started version: " + version)
	stats.Incr("started", nil, 1)

	handlers := createFaaSHandlers(nomadClient, consulResolver, stats, logger)

	config := &types.FaaSConfig{}
	config.ReadTimeout = *functionTimeout
	config.WriteTimeout = *functionTimeout
	config.TCPPort = port

	logger.Info("Started Nomad provider", "port", *config.TCPPort)
	bootstrap.Serve(handlers, config)
}

func createFaaSHandlers(nomadClient *api.Client, consulResolver *consul.Resolver, stats *statsd.Client, logger hclog.Logger) *types.FaaSHandlers {

	return &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(nomadClient.Jobs(), logger, stats),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs(), logger, stats),
		DeleteHandler:  handlers.MakeDelete(consulResolver, nomadClient.Jobs(), logger, stats),
		ReplicaReader:  makeReplicationReader(nomadClient.Jobs(), logger, stats),
		ReplicaUpdater: makeReplicationUpdater(nomadClient.Jobs(), logger, stats),
		FunctionProxy:  makeFunctionProxyHandler(consulResolver, logger, stats, *functionTimeout),
		UpdateHandler:  handlers.MakeDeploy(nomadClient.Jobs(), logger, stats),
	}
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
	logger.Info("create nomad client", "addr", nomadAddr)
	nomadClient, err := api.NewClient(c.ClientConfig(region, nomadAddr, false))
	if err != nil {
		logger.Error("Unable to create nomad client", err)
	}

	cr := consul.NewResolver(consulAddr, logger.Named("consul_resolver"))

	return logger, stats, nomadClient, cr
}

func setupLogging() hclog.Logger {
	logJSON := false
	if *loggerFormat == "json" {
		logJSON = true
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:       "nomadd",
		Level:      hclog.LevelFromString(*loggerLevel),
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
func makeFunctionProxyHandler(r consul.ServiceResolver, logger hclog.Logger, s *statsd.Client, timeout time.Duration) http.HandlerFunc {
	return handlers.MakeExtractFunctionMiddleWare(
		func(r *http.Request) map[string]string {
			return mux.Vars(r)
		},
		handlers.MakeProxy(
			handlers.ProxyConfig{
				Client:   handlers.MakeProxyClient(timeout, logger),
				Resolver: r,
				Logger:   logger,
				StatsD:   s,
				Timeout:  timeout,
			},
		),
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
