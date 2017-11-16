package main

import (
	"net/http"

	"github.com/nicholasjackson/bench/plugin/shared"
)

// BenchImpl implements the shared.Bench interface for bench plugins
type BenchImpl struct{}

// Do performs work associated with the benchmark run, it is called from
// the main bench thread
// e.g.
// func(b *BenchImpl) Do() error {
//  	resp, err := http.Get("http://www.amazon.co.uk/")
//  	defer func(response *http.Response) {
//  		if response != nil && response.Body != nil {
//  			response.Body.Close()
//  		}
//  	}(resp)
//
//  	if err != nil || resp.StatusCode != 200 {
//  		return err
//  	}
//
//  	return nil
// }
func (b *BenchImpl) Do() error {
	// perform any required work here

	// return an error if the work is not successful
	// return nil on success
	//resp, err := http.Post("http://192.168.1.113:8080/function/info", "", nil)
	resp, err := http.Post("http://openfaas-openfaas-689108950.eu-west-1.elb.amazonaws.com:8080/function/info", "", nil)
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
	// start the plugin
	shared.RunPlugin(&BenchImpl{})
}
