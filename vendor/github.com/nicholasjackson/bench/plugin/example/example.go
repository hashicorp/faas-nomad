package main

import (
	"net/http"

	"github.com/nicholasjackson/bench/plugin/shared"
)

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
