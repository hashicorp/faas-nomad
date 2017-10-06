# Bench
DOCUMENTATION WIP 18/08/2017

[![GoDoc](https://godoc.org/github.com/nicholasjackson/bench?status.svg)](https://godoc.org/github.com/nicholasjackson/bench) 
[![CircleCI](https://circleci.com/gh/nicholasjackson/bench.svg?style=svg)](https://circleci.com/gh/nicholasjackson/bench)
[![Coverage Status](https://coveralls.io/repos/github/nicholasjackson/bench/badge.svg?branch=master)](https://coveralls.io/github/nicholasjackson/bench?branch=master)

Bench is a simple load testing application for Microservices, I built this as I wanted to be able run some basic performance tests against the microservice frameworks which I am evaluating for my book "Building Microservices In Go".  Whilst there are many options to be able to test HTTP based services I could find nothing which would serve the purpose of being able to benchmark an RPC server or another framework which did not use JSON or another text based object like XML over REST.  

With bench you write your request as a simple Go application which implements bench's plugin format and then by configuring the run options such as the number of threads, timeout and run time, Bench will execute the given function and collect the results.  These results can then be processed by one or more of the given output writers to be able to output simple tabular summaries or more detailed charts.

Because Bench executes code compiled into a plugin, it is possible to use bench in any situation where you would like to test any code in a concurrent manner.

# Installing bench
```bash
$ go get -u github.com/nicholasjackson/bench/cli
```

# Commands
```
$bench -h
Usage: bench [--version] [--help] <command> [<args>]

Available commands are:
    init    scaffold a new plugin
    run     run the benchmarks
```

## Run
```
 $bench run -h
Usage of run:
  -duration duration
        the duration of the test e.g. 5s (5 seconds) (default 10s)
  -plugin string
        specify the location of the bench plugin
  -ramp duration
        time taken to schedule maximum threads e.g. 5s (5 seconds) (default 10s)
  -thread int
        the number of concurrent threads when running a benchmark (default 1)
  -timeout duration
        timeout value for a test e.g. 5s (5 seconds) (default 5s)
 
```

## Init
```
Usage of init:
  -output string
        specify the location of the outputed plugin code
```

# How Bench works
Bench will call the plugin's `Do` function repeatedly for a given time period, every test execution runs in a separate go function so it is possible to execute multiple instances concurrently.  The number of threads which are executed at any one time is user specifiable and is limited to the resources on the machine running the test.  There is also a simple thread ramp up algorythim which will increase the number of threads in a linear manner over a given time period.  

# Example
To create a new plugin make a new folder in your GOPATH and run the command:
```bash
$ bench init -output /OUTPUT_PATH/filename.go
```

This will create a new plugin in the given folder.

The basic example which can be found in the ./plugin/example directory contains a test to examine the performance of the amazon.co.uk home page:
```go
// BenchImpl implements shared.Bench interface
type BenchImpl struct{}

// Do executes a request and returns error
func (b BenchImpl) Do() error {
	resp, err := http.Get("http://www.amazon.co.uk/")
	defer func(response *http.Response) {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}(resp)

	if err != nil || resp.StatusCode != 200 {
		return err
	}

	return nil
}

func main() {
	shared.RunPlugin(&BenchImpl{})
}
```

Bench plugins are executable go programs so before we can use our plugin we need to compile it:
```bash
$ go build -o example example.go
```

We can now run the bench command and pass it parameters for:
* max threads
* test duration
* thread ramp up
* request timeout
* plugin location we have just compiled

Running bench is as simple as:
```bash
$ bench run -plugin path/to/plugin -threads 10 -duration 10s -ramp 0s -timeout 5s

#...

Waiting for threads to finish .......... OK

Start Time: 2017-08-18 13:48:14.517907 +0000 UTC
Elapsed Time: 10s
Threads: 10
Total Requests: 1194
Avg Request Time: 83.95ms
Total Success: 1194
Total Timeouts: 0
Total Failures: 0
```

Bench will also save a text file of the test and a simple chart to the folder from where you ran the `bench` command:

![](https://raw.githubusercontent.com/nicholasjackson/bench/master/example/output.png)

This chart is useful to see the response time of the given function against the current number of requests executed.  When load testing an API it is common to see the request time increase as the number of executed requests increases.

TODO:  
[ ] Update documentation  
[ ] Increase test coverage  
[ ] Write CSV data output  
