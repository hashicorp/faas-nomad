package main

import (
	"flag"
	"fmt"
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
	fntypes "github.com/hashicorp/faas-nomad/types"
	"github.com/hashicorp/faas-nomad/vault"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/mitchellh/mapstructure"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/types"
)

var version = "notset"

var (
	port                  = flag.Int("port", 8080, "Port to bind the server to")
	statsdServer          = flag.String("statsd_addr", "localhost:8125", "Location for the statsd collector")
	nodeURI               = flag.String("node_addr", "localhost", "URI of the current Nomad node, this address is used for reporting and logging")
	nomadAddr             = flag.String("nomad_addr", "localhost:4646", "Address for Nomad API endpoint")
	nomadTLSCA            = flag.String("nomad_tls_ca", "", "The TLS ca certificate file location")
	nomadTLSCert          = flag.String("nomad_tls_cert", "", "The TLS client certifcate file location")
	nomadTLSKey           = flag.String("nomad_tls_key", "", "The TLS private key file location")
	nomadTLSSkipVerify    = flag.Bool("nomad_tls_skip_verify", false, "Skips TLS verification for Nomad API. Not recommend for production")
	enableNomadTLS        = flag.Bool("enable_nomad_tls", false, "Toggles tls on Nomad endpoint/client")
	nomadACL              = flag.String("nomad_acl", "", "The ACL token for faas-nomad if Nomad ACLs are enabled")
	consulAddr            = flag.String("consul_addr", "http://localhost:8500", "Address for Consul API endpoint")
	consulACL             = flag.String("consul_acl", "", "ACL token for Consul API, only required if ACL are enabled in Consul")
	enableConsulDNS       = flag.Bool("enable_consul_dns", false, "Uses the consul_addr as a default DNS server. Assumes DNS interface is listening on port 53")
	nomadRegion           = flag.String("nomad_region", "global", "Default region to schedule functions in")
	enableBasicAuth       = flag.Bool("enable_basic_auth", false, "Flag for enabling basic authentication on gateway endpoints")
	basicAuthSecretPath   = flag.String("basic_auth_secret_path", "/secrets", "The directory path to the basic auth secret file")
	vaultAddrOverride     = flag.String("vault_addr", "", "Vault address override. Default Vault address is returned from the Nomad agent")
	vaultTLSSkipVerify    = flag.Bool("vault_tls_skip_verify", false, "Skips TLS verification for calls to Vault. Not recommend for production")
	vaultDefaultPolicy    = flag.String("vault_default_policy", "openfaas", "The default policy used when secrets are deployed with a function")
	vaultSecretPathPrefix = flag.String("vault_secret_path_prefix", "secret/openfaas", "The Vault k/v path prefix used when secrets are deployed with a function")
	vaultAppRoleID        = flag.String("vault_app_role_id", "", "A valid Vault AppRole role_id")
	vaultAppRoleSecretID  = flag.String("vault_app_secret_id", "", "A valid Vault AppRole secret_id derived from the role")
)

var functionTimeout = flag.Duration("function_timeout", 30*time.Second, "Timeout for function execution")

var (
	loggerFormat = flag.String("logger_format", "text", "Format for log output text | json")
	loggerLevel  = flag.String("logger_level", "INFO", "Log output level INFO | ERROR | DEBUG | TRACE")
	loggerOutput = flag.String("logger_output", "", "Filepath to write log file, if omitted stdOut is used")
)

func main() {
	flag.Parse()

	nomadConfig := &fntypes.NomadConfig{
		TLSEnabled:    *enableNomadTLS,
		Address:       *nomadAddr,
		ACLToken:      *nomadACL,
		TLSCA:         *nomadTLSCA,
		TLSCert:       *nomadTLSCert,
		TLSPrivateKey: *nomadTLSKey,
		TLSSkipVerify: *nomadTLSSkipVerify,
	}
	logger, stats, nomadClient, consulResolver := makeDependencies(
		*statsdServer,
		*nodeURI,
		*nomadConfig,
		*consulAddr,
		*consulACL,
		*nomadRegion,
	)

	logger.Info("Started version: " + version)
	stats.Incr("started", nil, 1)

	handlers := createFaaSHandlers(nomadClient, consulResolver, stats, logger)

	config := &types.FaaSConfig{}
	config.ReadTimeout = *functionTimeout
	config.WriteTimeout = *functionTimeout
	config.TCPPort = port
	config.EnableHealth = true
	config.EnableBasicAuth = *enableBasicAuth
	config.SecretMountPath = *basicAuthSecretPath

	logger.Info("Started Nomad provider", "port", *config.TCPPort)
	logger.Info("Basic authentication", "enabled", fmt.Sprintf("%t", config.EnableBasicAuth))

	bootstrap.Serve(handlers, config)
}

func createFaaSHandlers(nomadClient *api.Client, consulResolver *consul.Resolver, stats *statsd.Client, logger hclog.Logger) *types.FaaSHandlers {

	datacenter, err := nomadClient.Agent().Datacenter()
	if err != nil {
		logger.Error("Error returning the agent's datacenter", err.Error())
		datacenter = "dc1"
	}
	logger.Info("Datacenter from agent: " + datacenter)

	agentSelf, err := nomadClient.Agent().Self()
	var vaultConfig fntypes.VaultConfig
	if err != nil {
		logger.Error("/agent/self returned error. Unable to fetch Vault config.", err.Error())
	} else {
		mapstructure.Decode(agentSelf.Config["Vault"], &vaultConfig)
		if len(*vaultAddrOverride) > 0 {
			vaultConfig.Addr = *vaultAddrOverride
		}
	}

	logger.Info("Vault address: " + vaultConfig.Addr)
	vaultConfig.DefaultPolicy = *vaultDefaultPolicy
	vaultConfig.SecretPathPrefix = *vaultSecretPathPrefix
	vaultConfig.AppRoleID = *vaultAppRoleID
	vaultConfig.AppSecretID = *vaultAppRoleSecretID
	vaultConfig.TLSSkipVerify = *vaultTLSSkipVerify

	providerConfig := &fntypes.ProviderConfig{
		Vault:            vaultConfig,
		Datacenter:       datacenter,
		ConsulAddress:    *consulAddr,
		ConsulDNSEnabled: *enableConsulDNS,
	}

	vs := vault.NewVaultService(&vaultConfig, logger)

	_, loginErr := vs.Login()
	if loginErr != nil {
		logger.Error("Unable to login to Vault. Secrets will not work properly", loginErr.Error())
	} else {
		logger.Info("Vault authentication successful!")
	}

	return &types.FaaSHandlers{
		FunctionReader: handlers.MakeReader(nomadClient.Jobs(), logger, stats),
		DeployHandler:  handlers.MakeDeploy(nomadClient.Jobs(), *providerConfig, logger, stats),
		DeleteHandler:  handlers.MakeDelete(consulResolver, nomadClient.Jobs(), logger, stats),
		ReplicaReader:  makeReplicationReader(nomadClient.Jobs(), logger, stats),
		ReplicaUpdater: makeReplicationUpdater(nomadClient.Jobs(), logger, stats),
		FunctionProxy:  makeFunctionProxyHandler(consulResolver, logger, stats, *functionTimeout),
		UpdateHandler:  handlers.MakeDeploy(nomadClient.Jobs(), *providerConfig, logger, stats),
		InfoHandler:    handlers.MakeInfo(logger, stats, version),
		Health:         handlers.MakeHealthHandler(),
		SecretHandler:  handlers.MakeSecretHandler(vs, logger.Named("secrets_handler")),
	}
}

func makeDependencies(statsDAddr string, thisAddr string, nomadConfig fntypes.NomadConfig, consulAddr string, consulACL string, region string) (hclog.Logger, *statsd.Client, *api.Client, *consul.Resolver) {
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
	logger.Info("create nomad client", "addr", nomadConfig.Address)
	clientConfig := c.ClientConfig(region, nomadConfig.Address, nomadConfig.TLSEnabled)
	clientConfig.SecretID = nomadConfig.ACLToken
	if nomadConfig.TLSEnabled {
		clientConfig.TLSConfig = &api.TLSConfig{
			CACert:     nomadConfig.TLSCA,
			ClientCert: nomadConfig.TLSCert,
			ClientKey:  nomadConfig.TLSPrivateKey,
			Insecure:   nomadConfig.TLSSkipVerify,
		}
		clientConfig.ConfigureTLS()
	}

	nomadClient, err := api.NewClient(clientConfig)
	if err != nil {
		logger.Error("Unable to create nomad client", err)
	}

	cr := consul.NewResolver(consulAddr, consulACL, logger.Named("consul_resolver"))

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
