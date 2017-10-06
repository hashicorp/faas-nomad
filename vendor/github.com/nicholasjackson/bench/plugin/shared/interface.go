package shared

import (
	plugin "github.com/hashicorp/go-plugin"

	"github.com/nicholasjackson/bench/plugin/proto"
	"google.golang.org/grpc"
)

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"bench": &BenchPlugin{},
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BENCH_PLUGIN",
	MagicCookieValue: "hello",
}

type Bench interface {
	Do() error
}

// This is the implementation of plugin.Plugin so we can serve/consume this.
// We also implement GRPCPlugin so that this plugin can be served over
// gRPC.
type BenchPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Bench
}

func (p *BenchPlugin) GRPCServer(s *grpc.Server) error {
	proto.RegisterBenchPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *BenchPlugin) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewBenchPluginClient(c)}, nil
}
