package bench

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/nicholasjackson/bench/errors"
	"github.com/nicholasjackson/bench/results"
	"github.com/nicholasjackson/bench/semaphore"
)

// RunBenchmarks executes the benchmarks based upon the given criteria
//
// Returns a resultset
func (b *Bench) internalRun() results.ResultSet {

	startTime := time.Now()
	endTime := startTime.Add(b.duration)

	sem := semaphore.NewSemaphore(b.threads, b.rampUp) // create a new semaphore with an initiall capacity or 0
	out := make(chan results.Result)
	resultsChan := make(chan []results.Result)

	go handleResult(out, resultsChan)

	for run := true; run; run = (time.Now().Before(endTime)) {

		sem.Lock() // blocks when channel is full

		// execute a request
		go doRequest(b.request, b.timeout, sem, out)
	}

	fmt.Print("\nWaiting for threads to finish ")
	for i := sem.Length(); i != 0; i = sem.Length() {

		//abandon <- true
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println(" OK")
	fmt.Println("")

	close(out)

	return <-resultsChan
}

func handleResult(out chan results.Result, resultChan chan []results.Result) {

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	var r []results.Result

	for result := range out {
		switch result.Error.(type) {
		case errors.Timeout:
			yellow.Print("T")
		case error:
			red.Print("E")
		default:
			green.Print(".")
		}

		r = append(r, result)
	}

	resultChan <- r
}

func doRequest(r RequestFunc, timeout time.Duration, semaphore *semaphore.Semaphore, out chan results.Result) {

	// notify we are done at the end of the routine
	defer semaphore.Release()

	requestStart := time.Now()

	timeoutC := time.After(timeout)
	complete := make(chan results.Result)

	go func() {
		defer func() {
			if recover := recover(); r != nil {

				complete <- results.Result{
					Error:       fmt.Errorf("Panic: %v\n", recover),
					RequestTime: time.Now().Sub(requestStart),
					Timestamp:   time.Now(),
					Threads:     semaphore.Capacity(),
				}
			}
		}()

		err := r()
		complete <- results.Result{
			Error:       err,
			RequestTime: time.Now().Sub(requestStart),
			Timestamp:   time.Now(),
			Threads:     semaphore.Capacity(),
		}
	}()

	var ret results.Result

	select {
	case <-timeoutC:
		ret.Error = errors.Timeout{Message: "Timeout error"}
		ret.RequestTime = timeout
		ret.Timestamp = time.Now()
	case ret = <-complete:
	}
	out <- ret
}
