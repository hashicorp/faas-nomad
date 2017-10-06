package bench

import (
	"io"
	"time"

	"github.com/nicholasjackson/bench/output"
)

// RequestFunc defines a request that will be made by the
// benchmark system
type RequestFunc func() error

type outputContainer struct {
	interval time.Duration
	writer   io.Writer
	function output.OutputFunc
}

// Bench is the main application object it allows the profiling of calls to a remote endpoint defined by a Request
type Bench struct {
	threads  int
	duration time.Duration
	rampUp   time.Duration
	timeout  time.Duration
	outputs  []outputContainer
	request  RequestFunc
}

// New creates a new bench and intializes the intial values of...
// threads: the number of concurrent threads to execute
// duration: the duration to run the benchmarks for
// rampUp: the duration before the maximum number of threads is achieved
// timeout: the timeout value
// request: the Request to execute
//
// returns a new Bench instance
func New(
	threads int,
	duration time.Duration,
	rampUp time.Duration,
	timeout time.Duration) *Bench {

	return &Bench{
		threads:  threads,
		duration: duration,
		rampUp:   rampUp,
		timeout:  timeout,
	}
}

// AddOutput adds an output writer to Bench
func (b *Bench) AddOutput(interval time.Duration, writer io.Writer, output output.OutputFunc) {

	o := outputContainer{
		interval: interval,
		writer:   writer,
		function: output,
	}

	b.outputs = append(b.outputs, o)
}

// RunBenchmarks runs the benchmarking for the given function
func (b *Bench) RunBenchmarks(r RequestFunc) {

	b.request = r
	results := b.internalRun()
	b.processResults(results)
}
