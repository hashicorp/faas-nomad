package shared

import plugin "github.com/hashicorp/go-plugin"

// RunPlugin is a convenience method to start a bench plugin
func RunPlugin(impl Bench) {
	ps := map[string]plugin.Plugin{
		"bench": &BenchPlugin{Impl: impl},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         ps,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
