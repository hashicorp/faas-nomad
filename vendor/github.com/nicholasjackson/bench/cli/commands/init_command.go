package commands

import (
	"flag"
	"io/ioutil"
	"os"
)

var template = `
package main

import (
	"log"

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
	return nil
}

func main() {
	// start the plugin
	shared.RunPlugin(&BenchImpl{})
}
`

// Init is a cli command for initialising a new plugin
type Init struct {
	flagSet *flag.FlagSet

	outputLocation string
}

// Help returns the command help
func (i *Init) Help() string {
	i.flagSet.Usage()
	return ""
}

// Run runs the command
func (i *Init) Run(args []string) int {
	i.flagSet.Parse(args)

	if i.outputLocation != "" {
		createTemplate(i.outputLocation)
	}

	return 0
}

// Synopsis returns information about the command
func (i *Init) Synopsis() string {
	return "scaffold a new plugin"
}

// NewInit create a new Init command
func NewInit() *Init {
	i := &Init{
		flagSet:        flag.NewFlagSet("init", flag.ContinueOnError),
		outputLocation: "",
	}

	i.flagSet.StringVar(&i.outputLocation, "output", "", "specify the location of the outputed plugin code")

	return i
}

func createTemplate(location string) {
	ioutil.WriteFile(location, []byte(template), os.ModePerm)
}
