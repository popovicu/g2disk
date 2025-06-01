package protocol

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"libguestfs.org/nbdkit"

	"github.com/popovicu/g2disk/proto"
)

type G2DiskPluginConfig struct {
	endpoint string
}

type G2DiskGrpc struct {
	creds  credentials.TransportCredentials
	conn   *grpc.ClientConn
	client proto.G2DiskServiceClient
}

type G2DiskPlugin struct {
	nbdkit.Plugin

	logger *slog.Logger

	config G2DiskPluginConfig
	grpc   G2DiskGrpc
}

func NewUninitiatedPlugin(logger *slog.Logger) *G2DiskPlugin {
	return &G2DiskPlugin{
		logger: logger,
	}
}

type G2DiskConnection struct {
	nbdkit.Connection
	pluginHandle *G2DiskPlugin
}

func (p *G2DiskPlugin) Config(key string, value string) error {
	switch key {
	case "endpoint":
		p.config.endpoint = value
	default:
		return nbdkit.PluginError{Errmsg: fmt.Sprintf("unknown parameter %s", key)}
	}

	return nil
}

func (p *G2DiskPlugin) ConfigComplete() error {
	// TODO: verify configs

	if p.config.endpoint == "" {
		return fmt.Errorf("gRPC endpoint not configured")
	}

	p.logger.Info("Configuration complete")

	return nil
}

func (p *G2DiskPlugin) GetReady() error {
	p.logger.Info("Creating a gRPC client")
	var err error
	p.grpc.creds = insecure.NewCredentials()
	p.grpc.conn, err = grpc.NewClient(p.config.endpoint, grpc.WithTransportCredentials(p.grpc.creds))

	if err != nil {
		return fmt.Errorf("unable to set up a gRPC client: %v", err)
	}

	p.grpc.client = proto.NewG2DiskServiceClient(p.grpc.conn)
	return nil
}

func (p *G2DiskPlugin) Open(readonly bool) (nbdkit.ConnectionInterface, error) {
	return &G2DiskConnection{
		pluginHandle: p,
	}, nil
}

func (p *G2DiskPlugin) Unload() {
	p.logger.Info("Unloading the plugin, shutting down the gRPC client")
	defer p.grpc.conn.Close()
}

func (c *G2DiskConnection) GetSize() (uint64, error) {
	c.pluginHandle.logger.Info("Getting the disk size")
	req := &proto.GetSizeRequest{}
	resp, err := c.pluginHandle.grpc.client.GetSize(context.Background(), req)

	if err != nil {
		return 0, fmt.Errorf("unable to get the disk size: %v", err)
	}

	return resp.Size, nil
}

func (c *G2DiskConnection) CanMultiConn() (bool, error) {
	// Clients are not allowed to make multiple connections safely.
	// TODO: reconsider this
	return false, nil
}

func (c *G2DiskConnection) PRead(buf []byte, offset uint64,
	flags uint32) error {
	// TODO: consider the flags
	readSize := len(buf)
	req := &proto.ReadRequest{
		Offset:   offset,
		ReadSize: uint64(readSize),
	}

	resp, err := c.pluginHandle.grpc.client.Read(context.Background(), req)
	if err != nil {
		return fmt.Errorf("unable to read the disk at offset: %d, length: %d - %v", offset, readSize, err)
	}

	copy(buf, resp.Payload)
	return nil
}

func (c *G2DiskConnection) CanWrite() (bool, error) {
	return true, nil
}

func (c *G2DiskConnection) PWrite(buf []byte, offset uint64,
	flags uint32) error {
	req := &proto.WriteRequest{
		Offset:  offset,
		Payload: buf,
	}

	_, err := c.pluginHandle.grpc.client.Write(context.Background(), req)
	if err != nil {
		return fmt.Errorf("unable to write the disk at offset %d: %v", offset, err)
	}

	return nil
}
