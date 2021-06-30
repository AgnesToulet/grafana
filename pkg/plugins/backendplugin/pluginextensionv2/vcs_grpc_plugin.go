package pluginextensionv2

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type VCSPlugin interface {
	VersionedStorageClient
}

type VersionedStorageGRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

func (p *VersionedStorageGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	return nil
}

func (p *VersionedStorageGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &VersionedStorageGRPCClient{NewVersionedStorageClient(c)}, nil
}

type VersionedStorageGRPCClient struct {
	VersionedStorageClient
}

var _ plugin.GRPCPlugin = &VersionedStorageGRPCPlugin{}
