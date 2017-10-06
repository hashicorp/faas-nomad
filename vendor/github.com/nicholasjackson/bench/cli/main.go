package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/nicholasjackson/bench/cli/commands"
)

func main() {
	c := cli.NewCLI("bench", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"run":  runCommandFactory,
		"init": initCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

func runCommandFactory() (cli.Command, error) {
	return commands.NewRun(), nil
}

func initCommandFactory() (cli.Command, error) {
	return commands.NewInit(), nil
}
