package main

import (
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/nicholasjackson/bench/plugin/shared"
)

func TestStartsPlugin(t *testing.T) {
	ps := map[string]plugin.Plugin{
		"bench": &shared.BenchPlugin{Impl: &BenchImpl{}},
	}
	client, _ := plugin.TestPluginGRPCConn(t, ps)

	p, err := client.Dispense("bench")
	if err != nil {
		t.Fatal(err)
	}

	b := p.(shared.Bench)
	b.Do()
}
