package shared

import (
	context "golang.org/x/net/context"

	"github.com/nicholasjackson/bench/plugin/proto"
)

type GRPCServer struct {
	Impl Bench
}

func (m *GRPCServer) Do(ctx context.Context, req *proto.BenchRequest) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.Do()
}

type GRPCClient struct {
	client proto.BenchPluginClient
}

func (m *GRPCClient) Do() error {
	_, err := m.client.Do(context.Background(), &proto.BenchRequest{})
	return err
}
