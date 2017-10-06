package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/nicholasjackson/bench"
	"github.com/nicholasjackson/bench/output"
	"github.com/nicholasjackson/bench/plugin/shared"
	"github.com/nicholasjackson/bench/util"
)

// Run is a cli command which allows running of benchmarks
type Run struct {
	flagSet *flag.FlagSet

	pluginLocation string
	threads        int
	duration       time.Duration
	rampUp         time.Duration
	timeout        time.Duration

	client *plugin.Client
}

// Help returns the command help
func (r *Run) Help() string {
	r.flagSet.Usage()
	return ""
}

// Run runs the command
func (r *Run) Run(args []string) int {
	r.flagSet.Parse(args)

	if r.pluginLocation != "" {
		c, bp := createPlugin(r.pluginLocation)
		defer c.Kill()
		r.runBench(bp)
	}

	return 0
}

// Synopsis returns information about the command
func (r *Run) Synopsis() string {
	return "run the benchmarks"
}

// NewRun creates a new Run command
func NewRun() *Run {
	r := &Run{
		flagSet: flag.NewFlagSet("run", flag.ContinueOnError),
	}

	r.flagSet.StringVar(&r.pluginLocation, "plugin", "", "specify the location of the bench plugin")
	r.flagSet.IntVar(&r.threads, "thread", 1, "the number of concurrent threads when running a benchmark")
	r.flagSet.DurationVar(&r.duration, "duration", 10*time.Second, "the duration of the test e.g. 5s (5 seconds)")
	r.flagSet.DurationVar(&r.rampUp, "ramp", 10*time.Second, "time taken to schedule maximum threads e.g. 5s (5 seconds)")
	r.flagSet.DurationVar(&r.timeout, "timeout", 5*time.Second, "timeout value for a test e.g. 5s (5 seconds)")

	return r
}

func createPlugin(pluginLocation string) (*plugin.Client, shared.Bench) {
	logger := hclog.New(&hclog.LoggerOptions{
		Output: hclog.DefaultOutput,
		Level:  hclog.Info,
		Name:   "plugin",
	})

	c := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", pluginLocation),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	// Connect via RPC
	grpcClient, err := c.Client()
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	plug, err := grpcClient.Dispense("bench")
	if err != nil {
		fmt.Println("Error getting plugin:", err.Error())
		os.Exit(1)
	}

	return c, plug.(shared.Bench)
}

func (r *Run) runBench(bp shared.Bench) {
	b := bench.New(r.threads, r.duration, r.rampUp, r.timeout)

	b.AddOutput(0*time.Second, os.Stdout, output.WriteTabularData)
	b.AddOutput(1*time.Second, util.NewFile("./output.txt"), output.WriteTabularData)
	b.AddOutput(1*time.Second, util.NewFile("./output.png"), output.PlotData)
	b.AddOutput(0*time.Second, util.NewFile("./error.txt"), output.WriteErrorLogs)

	b.RunBenchmarks(func() error {
		return bp.Do()
	})
}
